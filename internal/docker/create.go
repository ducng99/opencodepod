package docker

import (
	"context"
	"fmt"
	"io"
	"net/netip"

	"opencodepod/internal/project"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) CreateProject(ctx context.Context, req *project.CreateRequest) (*project.Project, error) {
	id := generateID(8)
	image := req.Image
	if image == "" {
		image = dm.Cfg.DefaultImage
	}
	containerUser := req.ContainerUser
	if containerUser == "" {
		containerUser = dm.Cfg.ContainerUser
	}

	p := &project.Project{
		ID:            id,
		Name:          req.Name,
		Git:           project.Git{Repo: req.Git.Repo, Branch: req.Git.Branch, Depth: req.Git.Depth},
		Image:         image,
		Volumes:       project.ProjectVolumes(id),
		Status:        "creating",
		ContainerUser: containerUser,
	}

	for _, vol := range p.Volumes {
		_, err := dm.Client.VolumeCreate(ctx, dockerclient.VolumeCreateOptions{
			Name:   vol,
			Driver: "local",
			Labels: map[string]string{
				project.LabelManaged:   "true",
				project.LabelProjectID: id,
				project.LabelName:      req.Name,
			},
		})
		if err != nil {
			for _, v := range p.Volumes {
				_, _ = dm.Client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
			}
			return nil, fmt.Errorf("volume create: %w", err)
		}
	}

	// Try to pull the latest image; if that fails, fall back to a locally cached copy.
	pr, err := dm.Client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if _, inspectErr := dm.Client.ImageInspect(ctx, image); inspectErr != nil {
			for _, v := range p.Volumes {
				_, _ = dm.Client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
			}
			return nil, fmt.Errorf("image pull failed and no local image found: %w", err)
		}
	}

	// Port bindings: let Docker assign random host ports
	portBindings := network.PortMap{
		network.MustParsePort("22/tcp"):   []network.PortBinding{{HostIP: netip.IPv4Unspecified(), HostPort: "0"}},
		network.MustParsePort("8080/tcp"): []network.PortBinding{{HostIP: netip.IPv4Unspecified(), HostPort: "0"}},
	}

	// Exposed ports must be declared in Config
	exposedPorts := network.PortSet{
		network.MustParsePort("22/tcp"):   struct{}{},
		network.MustParsePort("8080/tcp"): struct{}{},
	}

	containerConfig := &container.Config{
		Image:        image,
		Labels:       project.LabelsFromProject(p),
		ExposedPorts: exposedPorts,
		Env:          dm.buildEnv(p),
	}

	extraHosts := []string{"host.docker.internal:host-gateway"}
	for host, ip := range dm.Cfg.Hosts {
		extraHosts = append(extraHosts, fmt.Sprintf("%s:%s", host, ip))
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        dm.buildBinds(p.ID, p.ContainerUser),
		ExtraHosts:   extraHosts,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	createResult, err := dm.Client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       project.ContainerName(id),
	})
	if err != nil {
		for _, v := range p.Volumes {
			_, _ = dm.Client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
		}
		return nil, fmt.Errorf("container create: %w", err)
	}

	if err := dm.injectSecrets(ctx, createResult.ID, p.ContainerUser); err != nil {
		return nil, err
	}

	if err := dm.startContainer(ctx, createResult.ID); err != nil {
		for _, v := range p.Volumes {
			_, _ = dm.Client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
		}
		return nil, err
	}

	inspectResult, err := dm.Client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect := inspectResult.Container

	populateProjectFromInspect(p, &inspect)

	return p, nil
}

package docker

import (
	"context"
	"errors"
	"fmt"
	"io"

	"opencodepod/internal/project"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) UpgradeProject(ctx context.Context, id string) (*project.Project, error) {
	inspectResult, err := dm.Client.ContainerInspect(ctx, project.ContainerName(id), dockerclient.ContainerInspectOptions{})
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, err
	}
	inspect := inspectResult.Container
	image := inspect.Config.Image

	// Capture the container's actual image ID so we can tell whether the pull
	// actually changed anything. We must use inspect.Image (the ID the container
	// was created from) rather than ImageInspect by tag, because the local tag
	// may already point to a newer image while the container still runs the old one.
	oldImageID := inspect.Image

	// Pull the latest image. The daemon streams JSON progress messages; we must
	// drain the entire body so the HTTP connection completes and the pull finishes.
	pr, err := dm.Client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if oldImageID == "" {
			return nil, fmt.Errorf("image pull failed and no local image found: %w", err)
		}
		return nil, fmt.Errorf("image pull failed; container was not upgraded because no newer image could be retrieved: %w", err)
	}

	// Check whether the image actually changed.
	newInspect, err := dm.Client.ImageInspect(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("image inspect after pull: %w", err)
	}
	if oldImageID != "" && oldImageID == newInspect.ID {
		// Same image — no need to recreate the container.
		return dm.GetProject(ctx, id)
	}

	// Newer image obtained. Stop and remove the old container (volume is preserved).
	if inspect.State.Status == "running" {
		if _, err := dm.Client.ContainerStop(ctx, inspect.ID, dockerclient.ContainerStopOptions{}); err != nil {
			return nil, fmt.Errorf("stop: %w", err)
		}
	}

	if _, err := dm.Client.ContainerRemove(ctx, inspect.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
		return nil, fmt.Errorf("remove old container: %w", err)
	}

	containerConfig := &container.Config{
		Image:        image,
		Labels:       inspect.Config.Labels,
		ExposedPorts: inspect.Config.ExposedPorts,
		Env:          inspect.Config.Env,
	}

	hostConfig := &container.HostConfig{
		PortBindings:  inspect.HostConfig.PortBindings,
		Binds:         inspect.HostConfig.Binds,
		ExtraHosts:    inspect.HostConfig.ExtraHosts,
		RestartPolicy: inspect.HostConfig.RestartPolicy,
	}

	createResult, err := dm.Client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       project.ContainerName(id),
	})
	if err != nil {
		return nil, fmt.Errorf("container create: %w", err)
	}

	if dm.Cfg.Git.Auth.SSHKey != "" {
		if err := dm.copyGitSSHKey(ctx, createResult.ID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git ssh key: %w", err)
		}
	}

	if dm.Cfg.Git.GPG.PrivateKey != "" {
		if err := dm.copyGPGKey(ctx, createResult.ID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy gpg key: %w", err)
		}
	}

	if len(dm.Cfg.Git.Auth.Credentials) > 0 {
		if err := dm.copyGitCredentials(ctx, createResult.ID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git credentials: %w", err)
		}
	}

	if _, err := dm.Client.ContainerStart(ctx, createResult.ID, dockerclient.ContainerStartOptions{}); err != nil {
		_, _ = dm.Client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("container start: %w", err)
	}

	inspectResult, err = dm.Client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect = inspectResult.Container

	p := project.ProjectFromLabels(id, inspect.Config.Labels)
	p.Image = image
	p.Volumes = project.ProjectVolumes(id)

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")

	return p, nil
}

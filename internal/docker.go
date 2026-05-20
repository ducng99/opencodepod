package internal

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"strconv"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	dockerclient "github.com/moby/moby/client"
)

type DockerManager struct {
	client *dockerclient.Client
	cfg    *Config
}

func NewDockerManager(cfg *Config) (*DockerManager, error) {
	cli, err := dockerclient.New(dockerclient.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerManager{client: cli, cfg: cfg}, nil
}

func (dm *DockerManager) Close() error {
	return dm.client.Close()
}

func (dm *DockerManager) ListProjects(ctx context.Context) ([]*Project, error) {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=true", LabelManaged))

	result, err := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0, len(result.Items))
	for _, c := range result.Items {
		p := dm.containerToProject(&c)
		projects = append(projects, p)
	}
	return projects, nil
}

func (dm *DockerManager) GetProject(ctx context.Context, id string) (*Project, error) {
	inspectResult, err := dm.client.ContainerInspect(ctx, ContainerName(id), dockerclient.ContainerInspectOptions{})
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, err
	}
	inspect := inspectResult.Container

	p := ProjectFromLabels(id, inspect.Config.Labels)
	p.Status = string(inspect.State.Status)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")
	return p, nil
}

func (dm *DockerManager) CreateProject(ctx context.Context, req *CreateRequest) (*Project, error) {
	id := generateID(8)
	image := req.Image
	if image == "" {
		image = dm.cfg.DefaultImage
	}

	p := &Project{
		ID:      id,
		Name:    req.Name,
		GitRepo: req.GitRepo,
		Image:   image,
		Volume:  VolumeName(id),
		Status:  "creating",
	}

	// Create volume
	volResult, err := dm.client.VolumeCreate(ctx, dockerclient.VolumeCreateOptions{
		Name:   p.Volume,
		Driver: "local",
		Labels: map[string]string{
			LabelManaged:   "true",
			LabelProjectID: id,
			LabelName:      req.Name,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("volume create: %w", err)
	}
	_ = volResult.Volume.Name

	// Try to pull the latest image; if that fails, fall back to a locally cached copy.
	pr, err := dm.client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if _, inspectErr := dm.client.ImageInspect(ctx, image); inspectErr != nil {
			_, _ = dm.client.VolumeRemove(ctx, p.Volume, dockerclient.VolumeRemoveOptions{Force: true})
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

	env := []string{}
	if req.GitRepo != "" {
		env = append(env, fmt.Sprintf("GIT_REPO=%s", req.GitRepo))
	}
	if dm.cfg.SSHPublicKey != "" {
		env = append(env, fmt.Sprintf("SSH_PUBLIC_KEY=%s", dm.cfg.SSHPublicKey))
	}

	containerConfig := &container.Config{
		Image:        image,
		Labels:       LabelsFromProject(p),
		ExposedPorts: exposedPorts,
		Env:          env,
	}

	binds := []string{fmt.Sprintf("%s:/workspace", p.Volume)}
	for _, m := range dm.cfg.Mounts {
		if m.Source == "" || m.Target == "" {
			continue
		}
		mode := "ro"
		if !m.ReadOnly {
			mode = "rw"
		}
		binds = append(binds, fmt.Sprintf("%s:%s:%s", m.Source, m.Target, mode))
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	createResult, err := dm.client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       ContainerName(id),
	})
	if err != nil {
		_, _ = dm.client.VolumeRemove(ctx, p.Volume, dockerclient.VolumeRemoveOptions{Force: true})
		return nil, fmt.Errorf("container create: %w", err)
	}

	if _, err := dm.client.ContainerStart(ctx, createResult.ID, dockerclient.ContainerStartOptions{}); err != nil {
		_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
		_, _ = dm.client.VolumeRemove(ctx, p.Volume, dockerclient.VolumeRemoveOptions{Force: true})
		return nil, fmt.Errorf("container start: %w", err)
	}

	// Inspect to get actual ports
	inspectResult, err := dm.client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect := inspectResult.Container

	p.Status = string(inspect.State.Status)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")

	return p, nil
}

func (dm *DockerManager) StartProject(ctx context.Context, id string) (*Project, error) {
	p, err := dm.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status == "running" {
		return p, nil
	}

	if _, err := dm.client.ContainerStart(ctx, ContainerName(id), dockerclient.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	return dm.GetProject(ctx, id)
}

func (dm *DockerManager) StopProject(ctx context.Context, id string) (*Project, error) {
	p, err := dm.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != "running" {
		return p, nil
	}

	if _, err := dm.client.ContainerStop(ctx, ContainerName(id), dockerclient.ContainerStopOptions{}); err != nil {
		return nil, fmt.Errorf("stop: %w", err)
	}
	return dm.GetProject(ctx, id)
}

func (dm *DockerManager) DeleteProject(ctx context.Context, id string) error {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	result, err := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return err
	}
	for _, c := range result.Items {
		if _, err := dm.client.ContainerRemove(ctx, c.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
			return fmt.Errorf("remove container: %w", err)
		}
	}
	if _, err := dm.client.VolumeRemove(ctx, VolumeName(id), dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
		// Volume may already be removed with container
		if !errors.Is(err, errdefs.ErrNotFound) {
			return fmt.Errorf("remove volume: %w", err)
		}
	}
	return nil
}

func (dm *DockerManager) containerToProject(c *container.Summary) *Project {
	p := ProjectFromLabels(c.Labels[LabelProjectID], c.Labels)
	p.Status = string(c.State)
	if p.Status == "" {
		p.Status = c.Status
	}

	// Parse ports from container summary (fast path)
	for _, port := range c.Ports {
		if port.PrivatePort == 22 {
			p.SSHPort = int(port.PublicPort)
		}
		if port.PrivatePort == 8080 {
			p.WebPort = int(port.PublicPort)
		}
	}
	return p
}

func (dm *DockerManager) refreshPorts(ctx context.Context, p *Project) (*Project, error) {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, p.ID))
	result, err := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("project not found: %s", p.ID)
	}
	inspectResult, err := dm.client.ContainerInspect(ctx, result.Items[0].ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}
	inspect := inspectResult.Container
	p.Status = string(inspect.State.Status)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")
	return p, nil
}

func hostPortFromInspect(inspect *container.InspectResponse, portKey string) int {
	bindings, ok := inspect.NetworkSettings.Ports[network.MustParsePort(portKey)]
	if !ok || len(bindings) == 0 {
		return 0
	}
	port, _ := strconv.Atoi(bindings[0].HostPort)
	return port
}

func generateID(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		v, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[v.Int64()]
	}
	return string(b)
}

package internal

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerManager struct {
	client *dockerclient.Client
	cfg    *Config
}

func NewDockerManager(cfg *Config) (*DockerManager, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerManager{client: cli, cfg: cfg}, nil
}

func (dm *DockerManager) Close() error {
	return dm.client.Close()
}

func (dm *DockerManager) ListProjects(ctx context.Context) ([]*Project, error) {
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=true", LabelManaged))

	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}

	projects := make([]*Project, 0, len(containers))
	for _, c := range containers {
		p := dm.containerToProject(&c)
		projects = append(projects, p)
	}
	return projects, nil
}

func (dm *DockerManager) GetProject(ctx context.Context, id string) (*Project, error) {
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))

	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return dm.containerToProject(&containers[0]), nil
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
	_, err := dm.client.VolumeCreate(ctx, volume.CreateOptions{
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

	// Port bindings: let Docker assign random host ports
	portBindings := nat.PortMap{
		"22/tcp":   []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "0"}},
		"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "0"}},
	}

	// Exposed ports must be declared in Config
	exposedPorts := nat.PortSet{
		"22/tcp":   struct{}{},
		"8080/tcp": struct{}{},
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

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        []string{fmt.Sprintf("%s:/home/coder/workspace", p.Volume)},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	resp, err := dm.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, ContainerName(id))
	if err != nil {
		_ = dm.client.VolumeRemove(ctx, p.Volume, true)
		return nil, fmt.Errorf("container create: %w", err)
	}

	if err := dm.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = dm.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		_ = dm.client.VolumeRemove(ctx, p.Volume, true)
		return nil, fmt.Errorf("container start: %w", err)
	}

	// Inspect to get actual ports
	inspect, err := dm.client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}

	p.Status = inspect.State.Status
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
		return dm.refreshPorts(ctx, p)
	}

	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	if err := dm.client.ContainerStart(ctx, containers[0].ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	return dm.refreshPorts(ctx, p)
}

func (dm *DockerManager) StopProject(ctx context.Context, id string) (*Project, error) {
	p, err := dm.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != "running" {
		return p, nil
	}

	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	if err := dm.client.ContainerStop(ctx, containers[0].ID, container.StopOptions{}); err != nil {
		return nil, fmt.Errorf("stop: %w", err)
	}
	return dm.GetProject(ctx, id)
}

func (dm *DockerManager) DeleteProject(ctx context.Context, id string) error {
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return err
	}
	for _, c := range containers {
		if err := dm.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			return fmt.Errorf("remove container: %w", err)
		}
	}
	if err := dm.client.VolumeRemove(ctx, VolumeName(id), true); err != nil {
		// Volume may already be removed with container
		if !dockerclient.IsErrNotFound(err) {
			return fmt.Errorf("remove volume: %w", err)
		}
	}
	return nil
}

func (dm *DockerManager) containerToProject(c *types.Container) *Project {
	p := ProjectFromLabels(c.Labels[LabelProjectID], c.Labels)
	p.Status = c.State
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
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, p.ID))
	containers, err := dm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("project not found: %s", p.ID)
	}
	inspect, err := dm.client.ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return nil, err
	}
	p.Status = inspect.State.Status
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")
	return p, nil
}

func hostPortFromInspect(inspect *types.ContainerJSON, portKey string) int {
	bindings, ok := inspect.NetworkSettings.Ports[nat.Port(portKey)]
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
		b[i] = letters[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(letters))]
	}
	return string(b)
}

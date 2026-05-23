package internal

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/netip"
	"net/url"
	"path/filepath"
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

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
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
		Volumes: ProjectVolumes(id),
		Status:  "creating",
	}

	for _, vol := range p.Volumes {
		_, err := dm.client.VolumeCreate(ctx, dockerclient.VolumeCreateOptions{
			Name:   vol,
			Driver: "local",
			Labels: map[string]string{
				LabelManaged:   "true",
				LabelProjectID: id,
				LabelName:      req.Name,
			},
		})
		if err != nil {
			for _, v := range p.Volumes {
				_, _ = dm.client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
			}
			return nil, fmt.Errorf("volume create: %w", err)
		}
	}

	// Try to pull the latest image; if that fails, fall back to a locally cached copy.
	pr, err := dm.client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if _, inspectErr := dm.client.ImageInspect(ctx, image); inspectErr != nil {
			for _, v := range p.Volumes {
				_, _ = dm.client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
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

	env := []string{}
	if req.GitRepo != "" {
		env = append(env, fmt.Sprintf("GIT_REPO=%s", req.GitRepo))
	}
	if dm.cfg.SSHPublicKey != "" {
		env = append(env, fmt.Sprintf("SSH_PUBLIC_KEY=%s", dm.cfg.SSHPublicKey))
	}
	if dm.cfg.Git.UserName != "" {
		env = append(env, fmt.Sprintf("GIT_USER_NAME=%s", dm.cfg.Git.UserName))
	}
	if dm.cfg.Git.UserEmail != "" {
		env = append(env, fmt.Sprintf("GIT_USER_EMAIL=%s", dm.cfg.Git.UserEmail))
	}
	if dm.cfg.Git.GPG.KeyID != "" {
		env = append(env, fmt.Sprintf("GIT_GPG_KEY_ID=%s", dm.cfg.Git.GPG.KeyID))
	}

	containerConfig := &container.Config{
		Image:        image,
		Labels:       LabelsFromProject(p),
		ExposedPorts: exposedPorts,
		Env:          env,
	}

	volumeTargets := []string{"/workspaces", "/home/coder/.local/share/opencode"}

	binds := make([]string, 0, len(p.Volumes)+len(dm.cfg.Mounts))
	for i, vol := range p.Volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", vol, volumeTargets[i]))
	}
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

	extraHosts := []string{"host.docker.internal:host-gateway"}
	for host, ip := range dm.cfg.Hosts {
		extraHosts = append(extraHosts, fmt.Sprintf("%s:%s", host, ip))
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		ExtraHosts:   extraHosts,
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
		for _, v := range p.Volumes {
			_, _ = dm.client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
		}
		return nil, fmt.Errorf("container create: %w", err)
	}

	// Inject Git SSH key directly into container filesystem before start.
	if dm.cfg.Git.Auth.SSHKey != "" {
		if err := dm.copyGitSSHKey(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git ssh key: %w", err)
		}
	}

	// Inject GPG private key directly into container filesystem before start.
	if dm.cfg.Git.GPG.PrivateKey != "" {
		if err := dm.copyGPGKey(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy gpg key: %w", err)
		}
	}

	// Inject Git HTTP credentials directly into container filesystem before start.
	if len(dm.cfg.Git.Auth.Credentials) > 0 {
		if err := dm.copyGitCredentials(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git credentials: %w", err)
		}
	}

	if _, err := dm.client.ContainerStart(ctx, createResult.ID, dockerclient.ContainerStartOptions{}); err != nil {
		_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
		for _, v := range p.Volumes {
			_, _ = dm.client.VolumeRemove(ctx, v, dockerclient.VolumeRemoveOptions{Force: true})
		}
		return nil, fmt.Errorf("container start: %w", err)
	}

	// Inspect to get actual ports
	inspectResult, err := dm.client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect := inspectResult.Container

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
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

func (dm *DockerManager) UpgradeProject(ctx context.Context, id string) (*Project, error) {
	inspectResult, err := dm.client.ContainerInspect(ctx, ContainerName(id), dockerclient.ContainerInspectOptions{})
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
	pr, err := dm.client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
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
	newInspect, err := dm.client.ImageInspect(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("image inspect after pull: %w", err)
	}
	if oldImageID != "" && oldImageID == newInspect.ID {
		// Same image — no need to recreate the container.
		return dm.GetProject(ctx, id)
	}

	// Newer image obtained. Stop and remove the old container (volume is preserved).
	if inspect.State.Status == "running" {
		if _, err := dm.client.ContainerStop(ctx, inspect.ID, dockerclient.ContainerStopOptions{}); err != nil {
			return nil, fmt.Errorf("stop: %w", err)
		}
	}

	if _, err := dm.client.ContainerRemove(ctx, inspect.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
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

	createResult, err := dm.client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       ContainerName(id),
	})
	if err != nil {
		return nil, fmt.Errorf("container create: %w", err)
	}

	if dm.cfg.Git.Auth.SSHKey != "" {
		if err := dm.copyGitSSHKey(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git ssh key: %w", err)
		}
	}

	if dm.cfg.Git.GPG.PrivateKey != "" {
		if err := dm.copyGPGKey(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy gpg key: %w", err)
		}
	}

	if len(dm.cfg.Git.Auth.Credentials) > 0 {
		if err := dm.copyGitCredentials(ctx, createResult.ID); err != nil {
			_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
			return nil, fmt.Errorf("copy git credentials: %w", err)
		}
	}

	if _, err := dm.client.ContainerStart(ctx, createResult.ID, dockerclient.ContainerStartOptions{}); err != nil {
		_, _ = dm.client.ContainerRemove(ctx, createResult.ID, dockerclient.ContainerRemoveOptions{Force: true})
		return nil, fmt.Errorf("container start: %w", err)
	}

	inspectResult, err = dm.client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect = inspectResult.Container

	p := ProjectFromLabels(id, inspect.Config.Labels)
	p.Image = image
	p.Volumes = ProjectVolumes(id)

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")

	return p, nil
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
	for _, vol := range ProjectVolumes(id) {
		if _, err := dm.client.VolumeRemove(ctx, vol, dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
			if !errors.Is(err, errdefs.ErrNotFound) {
				return fmt.Errorf("remove volume %s: %w", vol, err)
			}
		}
	}
	return nil
}

func computeStatus(state, healthStatus string) string {
	if state != string(container.StateRunning) {
		return "stopped"
	}
	switch healthStatus {
	case string(container.Starting):
		return "starting"
	case string(container.Unhealthy):
		return "unhealthy"
	default:
		return "running"
	}
}

func (dm *DockerManager) containerToProject(c *container.Summary) *Project {
	p := ProjectFromLabels(c.Labels[LabelProjectID], c.Labels)

	healthStatus := ""
	if c.Health != nil {
		healthStatus = string(c.Health.Status)
	}
	p.Status = computeStatus(string(c.State), healthStatus)

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

func (dm *DockerManager) refreshState(ctx context.Context, p *Project) (*Project, error) {
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

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
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

func (dm *DockerManager) copyGitSSHKey(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(dm.cfg.Git.Auth.SSHKey)
	hdr := &tar.Header{
		Name: filepath.Base(dm.cfg.Git.Auth.SSHKeyPath),
		Mode: 0o600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	_, err := dm.client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: filepath.Dir(dm.cfg.Git.Auth.SSHKeyPath),
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGPGKey(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(dm.cfg.Git.GPG.PrivateKey)
	hdr := &tar.Header{
		Name: ".gnupg/private.key",
		Mode: 0o600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	_, err := dm.client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: "/home/coder",
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGitCredentials(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	var content bytes.Buffer
	for host, cred := range dm.cfg.Git.Auth.Credentials {
		username := url.QueryEscape(cred.Username)
		password := url.QueryEscape(cred.Password)
		content.WriteString(fmt.Sprintf("https://%s:%s@%s\n", username, password, host))
	}

	hdr := &tar.Header{
		Name: ".git-credentials",
		Mode: 0o600,
		Size: int64(content.Len()),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(content.Bytes()); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	_, err := dm.client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: "/home/coder",
		Content:         &buf,
	})
	return err
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

package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"path/filepath"
	"strconv"

	"opencodepod/internal/config"
	"opencodepod/internal/project"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	dockerclient "github.com/moby/moby/client"
)

type DockerManager struct {
	Client *dockerclient.Client
	Cfg    *config.Config
}

func NewDockerManager(cfg *config.Config) (*DockerManager, error) {
	cli, err := dockerclient.New(dockerclient.FromEnv)
	if err != nil {
		return nil, err
	}
	return &DockerManager{Client: cli, Cfg: cfg}, nil
}

func (dm *DockerManager) Close() error {
	return dm.Client.Close()
}

func appendEnv(env []string, key, val string) []string {
	if val != "" {
		return append(env, key+"="+val)
	}
	return env
}

func appendEnvInt(env []string, key string, val int) []string {
	if val > 0 {
		return append(env, fmt.Sprintf("%s=%d", key, val))
	}
	return env
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

func (dm *DockerManager) containerToProject(c *container.Summary) *project.Project {
	p := project.ProjectFromLabels(c.Labels[project.LabelProjectID], c.Labels)

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

func (dm *DockerManager) refreshState(ctx context.Context, p *project.Project) (*project.Project, error) {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", project.LabelProjectID, p.ID))
	result, err := dm.Client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("project not found: %s", p.ID)
	}
	inspectResult, err := dm.Client.ContainerInspect(ctx, result.Items[0].ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}
	inspect := inspectResult.Container

	populateProjectFromInspect(p, &inspect)
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

func (dm *DockerManager) inspectProject(ctx context.Context, id string) (container.InspectResponse, error) {
	result, err := dm.Client.ContainerInspect(ctx, project.ContainerName(id), dockerclient.ContainerInspectOptions{})
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return container.InspectResponse{}, fmt.Errorf("project not found: %s", id)
		}
		return container.InspectResponse{}, err
	}
	return result.Container, nil
}

func populateProjectFromInspect(p *project.Project, inspect *container.InspectResponse) {
	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
	p.SSHPort = hostPortFromInspect(inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(inspect, "8080/tcp")
}

func (dm *DockerManager) stopAndRemoveContainer(ctx context.Context, containerID string) error {
	inspect, err := dm.Client.ContainerInspect(ctx, containerID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return err
	}
	if inspect.Container.State.Status == "running" {
		if _, err := dm.Client.ContainerStop(ctx, containerID, dockerclient.ContainerStopOptions{}); err != nil {
			return fmt.Errorf("stop: %w", err)
		}
	}
	if _, err := dm.Client.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("remove old container: %w", err)
	}
	return nil
}

func (dm *DockerManager) injectSecrets(ctx context.Context, containerID string) error {
	if dm.Cfg.Git.Auth.SSHKey != "" {
		if err := dm.copyGitSSHKey(ctx, containerID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{Force: true})
			return fmt.Errorf("copy git ssh key: %w", err)
		}
	}
	if dm.Cfg.Git.GPG.PrivateKey != "" {
		if err := dm.copyGPGKey(ctx, containerID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{Force: true})
			return fmt.Errorf("copy gpg key: %w", err)
		}
	}
	if len(dm.Cfg.Git.Auth.Credentials) > 0 {
		if err := dm.copyGitCredentials(ctx, containerID); err != nil {
			_, _ = dm.Client.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{Force: true})
			return fmt.Errorf("copy git credentials: %w", err)
		}
	}
	return nil
}

func (dm *DockerManager) startContainer(ctx context.Context, containerID string) error {
	if _, err := dm.Client.ContainerStart(ctx, containerID, dockerclient.ContainerStartOptions{}); err != nil {
		_, _ = dm.Client.ContainerRemove(ctx, containerID, dockerclient.ContainerRemoveOptions{Force: true})
		return fmt.Errorf("container start: %w", err)
	}
	return nil
}

func writeTarToContainer(ctx context.Context, client *dockerclient.Client, containerID, destPath, filename string, content []byte) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: filename,
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
	_, err := client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: destPath,
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGitSSHKey(ctx context.Context, containerID string) error {
	return writeTarToContainer(ctx, dm.Client, containerID, filepath.Dir(dm.Cfg.Git.Auth.SSHKeyPath), filepath.Base(dm.Cfg.Git.Auth.SSHKeyPath), []byte(dm.Cfg.Git.Auth.SSHKey))
}

func (dm *DockerManager) copyGPGKey(ctx context.Context, containerID string) error {
	return writeTarToContainer(ctx, dm.Client, containerID, "/home/coder", ".gnupg/private.key", []byte(dm.Cfg.Git.GPG.PrivateKey))
}

func (dm *DockerManager) copyGitCredentials(ctx context.Context, containerID string) error {
	var content bytes.Buffer
	for host, cred := range dm.Cfg.Git.Auth.Credentials {
		username := url.QueryEscape(cred.Username)
		password := url.QueryEscape(cred.Password)
		content.WriteString(fmt.Sprintf("https://%s:%s@%s\n", username, password, host))
	}
	return writeTarToContainer(ctx, dm.Client, containerID, "/home/coder", ".git-credentials", content.Bytes())
}

func (dm *DockerManager) buildEnv(p *project.Project) []string {
	env := []string{}
	env = appendEnv(env, "GIT_REPO", p.Git.Repo)
	env = appendEnv(env, "GIT_BRANCH", p.Git.Branch)
	env = appendEnvInt(env, "GIT_DEPTH", p.Git.Depth)
	env = appendEnv(env, "SSH_PUBLIC_KEY", dm.Cfg.SSHPublicKey)
	env = appendEnv(env, "GIT_USER_NAME", dm.Cfg.Git.UserName)
	env = appendEnv(env, "GIT_USER_EMAIL", dm.Cfg.Git.UserEmail)
	env = appendEnv(env, "GIT_GPG_KEY_ID", dm.Cfg.Git.GPG.KeyID)
	env = appendEnv(env, "GPG_PASSPHRASE_PATH", dm.Cfg.Git.GPG.PassphrasePath)
	return env
}

func (dm *DockerManager) buildBinds(id string) []string {
	binds := make([]string, 0, len(project.ProjectVolumeMounts(id))+len(dm.Cfg.Mounts))
	for _, mount := range project.ProjectVolumeMounts(id) {
		binds = append(binds, fmt.Sprintf("%s:%s", mount.Name, mount.Target))
	}
	for _, m := range dm.Cfg.Mounts {
		if m.Source == "" || m.Target == "" {
			continue
		}
		mode := "ro"
		if !m.ReadOnly {
			mode = "rw"
		}
		binds = append(binds, fmt.Sprintf("%s:%s:%s", m.Source, m.Target, mode))
	}
	return binds
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

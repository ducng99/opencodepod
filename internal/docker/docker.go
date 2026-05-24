package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"path/filepath"
	"strconv"

	"opencodepod/internal/config"
	"opencodepod/internal/project"

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

	content := []byte(dm.Cfg.Git.Auth.SSHKey)
	hdr := &tar.Header{
		Name: filepath.Base(dm.Cfg.Git.Auth.SSHKeyPath),
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

	_, err := dm.Client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: filepath.Dir(dm.Cfg.Git.Auth.SSHKeyPath),
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGPGKey(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(dm.Cfg.Git.GPG.PrivateKey)
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

	_, err := dm.Client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: "/home/coder",
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGPGPassphrase(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	content := []byte(dm.Cfg.Git.GPG.Passphrase)
	hdr := &tar.Header{
		Name: ".gnupg/gpg_passphrase.key",
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

	_, err := dm.Client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
		DestinationPath: "/home/coder",
		Content:         &buf,
	})
	return err
}

func (dm *DockerManager) copyGitCredentials(ctx context.Context, containerID string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	var content bytes.Buffer
	for host, cred := range dm.Cfg.Git.Auth.Credentials {
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

	_, err := dm.Client.CopyToContainer(ctx, containerID, dockerclient.CopyToContainerOptions{
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

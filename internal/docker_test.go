package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

// testImage is a lightweight image that stays running so we can test start/stop.
const testImage = "nginx:alpine"

func skipIfNoDocker(t *testing.T) *DockerManager {
	t.Helper()
	cfg := &Config{
		ListenAddr:   ":8080",
		DefaultImage: testImage,
	}
	dm, err := NewDockerManager(cfg)
	if err != nil {
		t.Skipf("docker client unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = dm.Close()
	})

	// Verify we can actually reach the daemon by pinging it.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := dm.client.Ping(ctx, dockerclient.PingOptions{}); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	// Ensure test image is present; pull if missing.
	if _, err := dm.client.ImageInspect(ctx, testImage); err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			pullCtx, pullCancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer pullCancel()
			pr, err := dm.client.ImagePull(pullCtx, testImage, dockerclient.ImagePullOptions{})
			if err != nil {
				t.Skipf("unable to pull test image %s: %v", testImage, err)
			}
			_, _ = io.Copy(io.Discard, pr)
			_ = pr.Close()
		} else {
			t.Skipf("unable to inspect test image %s: %v", testImage, err)
		}
	}

	return dm
}

func cleanupTestProject(t *testing.T, dm *DockerManager, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Remove container(s)
	f := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	result, err := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		t.Errorf("cleanup container list failed: %v", err)
		return
	}
	for _, c := range result.Items {
		if _, err := dm.client.ContainerRemove(ctx, c.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
			t.Errorf("cleanup container remove failed: %v", err)
		}
	}
	// Remove volumes
	for _, vol := range ProjectVolumes(id) {
		if _, err := dm.client.VolumeRemove(ctx, vol, dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
			t.Errorf("cleanup volume %s remove failed: %v", vol, err)
		}
	}
}

func TestDockerManager_ListProjects(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	projects, err := dm.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects failed: %v", err)
	}
	// We don't know how many projects exist outside our test; just ensure no panic.
	_ = projects
}

func TestDockerManager_CreateProject_PullsImage(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	// Remove the image if present so we can test auto-pull.
	_, _ = dm.client.ImageRemove(ctx, testImage, dockerclient.ImageRemoveOptions{Force: true})

	req := &CreateRequest{Name: "test-autopull", Image: testImage}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed to auto-pull image: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	if p.Status == "" {
		t.Error("expected non-empty status")
	}
}

func TestDockerManager_CreateProject(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	req := &CreateRequest{Name: "test-create"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected project ID")
	}
	defer cleanupTestProject(t, dm, p.ID)

	if p.Name != "test-create" {
		t.Errorf("expected name test-create, got %s", p.Name)
	}
	if len(p.Volumes) != 2 {
		t.Errorf("expected 2 volumes, got %d", len(p.Volumes))
	}
	if p.Volumes[0] != VolumeName(p.ID) {
		t.Errorf("expected volume %s, got %s", VolumeName(p.ID), p.Volumes[0])
	}
	if p.Volumes[1] != HomeVolumeName(p.ID) {
		t.Errorf("expected home volume %s, got %s", HomeVolumeName(p.ID), p.Volumes[1])
	}
	if p.Status == "" {
		t.Error("expected non-empty status")
	}
	if p.SSHPort == 0 {
		t.Error("expected a host port assigned for SSH")
	}
	if p.WebPort == 0 {
		t.Error("expected a host port assigned for web")
	}

	// Verify container exists with correct labels
	f := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, p.ID))
	result, err := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		t.Fatalf("container list failed: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 container, got %d", len(result.Items))
	}
	c := result.Items[0]
	if c.Labels[LabelProjectID] != p.ID {
		t.Errorf("expected container label project.id=%s", p.ID)
	}
	if c.Labels[LabelName] != "test-create" {
		t.Errorf("expected container label project.name=test-create")
	}

	// Verify volumes exist
	for _, vol := range p.Volumes {
		volResult, err := dm.client.VolumeInspect(ctx, vol, dockerclient.VolumeInspectOptions{})
		if err != nil {
			t.Fatalf("volume inspect failed for %s: %v", vol, err)
		}
		v := volResult.Volume
		if v.Name != vol {
			t.Errorf("expected volume name %s, got %s", vol, v.Name)
		}
	}
}

func TestDockerManager_GetProject(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	req := &CreateRequest{Name: "test-get"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	got, err := dm.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected id %s, got %s", p.ID, got.ID)
	}
	if got.Name != p.Name {
		t.Errorf("expected name %s, got %s", p.Name, got.Name)
	}

	// Not found
	_, err = dm.GetProject(ctx, "nonexistent")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestDockerManager_StopStartProject(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	req := &CreateRequest{Name: "test-lifecycle"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	// Stop
	stopped, err := dm.StopProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("StopProject failed: %v", err)
	}
	if stopped.Status == "running" {
		t.Errorf("expected stopped status, got %s", stopped.Status)
	}

	// Start
	started, err := dm.StartProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("StartProject failed: %v", err)
	}
	if started.Status != "running" {
		t.Errorf("expected running status after start, got %s", started.Status)
	}

	// Stop again is idempotent-ish
	stopped2, err := dm.StopProject(ctx, p.ID)
	if err != nil {
		t.Fatalf("second StopProject failed: %v", err)
	}
	if stopped2.Status == "running" {
		t.Errorf("expected stopped status, got %s", stopped2.Status)
	}
}

func TestDockerManager_DeleteProject(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	req := &CreateRequest{Name: "test-delete"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	id := p.ID
	if err := dm.DeleteProject(ctx, id); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify gone
	f := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", LabelProjectID, id))
	result, _ := dm.client.ContainerList(ctx, dockerclient.ContainerListOptions{All: true, Filters: f})
	if len(result.Items) > 0 {
		t.Errorf("expected 0 containers after delete, got %d", len(result.Items))
	}

	for _, vol := range ProjectVolumes(id) {
		_, err = dm.client.VolumeInspect(ctx, vol, dockerclient.VolumeInspectOptions{})
		if err == nil {
			t.Errorf("expected volume %s to be removed", vol)
		}
	}
}

func TestDockerManager_CreateProject_WithSSHKey(t *testing.T) {
	ctx := context.Background()

	cfgWithSSH := &Config{
		ListenAddr:   ":8080",
		DefaultImage: testImage,
		SSHPublicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test",
	}
	dm, err := NewDockerManager(cfgWithSSH)
	if err != nil {
		t.Skipf("docker client unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = dm.Close()
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := dm.client.Ping(pingCtx, dockerclient.PingOptions{}); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	req := &CreateRequest{Name: "test-ssh-key"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	// Inspect container to verify the SSH_PUBLIC_KEY env var is present.
	inspectResult, err := dm.client.ContainerInspect(ctx, ContainerName(p.ID), dockerclient.ContainerInspectOptions{})
	if err != nil {
		t.Fatalf("container inspect failed: %v", err)
	}
	inspect := inspectResult.Container

	expected := "SSH_PUBLIC_KEY=ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI test"
	found := false
	for _, e := range inspect.Config.Env {
		if e == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected env var %q not found in container env %v", expected, inspect.Config.Env)
	}
}

func TestDockerManager_CreateProject_WithGitSSHKey(t *testing.T) {
	ctx := context.Background()

	expectedKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI testkey"
	cfgWithKey := &Config{
		ListenAddr:   ":8080",
		DefaultImage: testImage,
		Git: GitConfig{
			Auth: GitAuthConfig{
				SSHKey:     expectedKey,
				SSHKeyPath: "/tmp/git_ssh_key",
			},
		},
	}
	dm, err := NewDockerManager(cfgWithKey)
	if err != nil {
		t.Skipf("docker client unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = dm.Close()
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := dm.client.Ping(pingCtx, dockerclient.PingOptions{}); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	req := &CreateRequest{Name: "test-git-ssh-key"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	// Verify the key file exists inside the container.
	execCreateResult, err := dm.client.ExecCreate(ctx, ContainerName(p.ID), dockerclient.ExecCreateOptions{
		AttachStdout: true,
		TTY:          true,
		Cmd:          []string{"cat", cfgWithKey.Git.Auth.SSHKeyPath},
	})
	if err != nil {
		t.Fatalf("container exec create failed: %v", err)
	}

	attachResult, err := dm.client.ExecAttach(ctx, execCreateResult.ID, dockerclient.ExecAttachOptions{TTY: true})
	if err != nil {
		t.Fatalf("container exec attach failed: %v", err)
	}
	defer attachResult.Close()

	data, err := io.ReadAll(attachResult.Reader)
	if err != nil {
		t.Fatalf("read exec output failed: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != expectedKey {
		t.Errorf("expected key file content %q, got %q", expectedKey, got)
	}
}

func TestDockerManager_containerToProject(t *testing.T) {
	dm := skipIfNoDocker(t)
	c := &container.Summary{
		ID:     "cid",
		Labels: map[string]string{LabelProjectID: "pid", LabelName: "n"},
		State:  container.StateRunning,
		Ports:  []container.PortSummary{{PrivatePort: 22, PublicPort: 10022}, {PrivatePort: 8080, PublicPort: 18080}},
	}
	p := dm.containerToProject(c)
	if p.ID != "pid" {
		t.Errorf("expected pid, got %s", p.ID)
	}
	if p.Status != "running" {
		t.Errorf("expected running, got %s", p.Status)
	}
	if p.SSHPort != 10022 {
		t.Errorf("expected ssh port 10022, got %d", p.SSHPort)
	}
	if p.WebPort != 18080 {
		t.Errorf("expected web port 18080, got %d", p.WebPort)
	}
}

func TestDockerManager_refreshPorts(t *testing.T) {
	dm := skipIfNoDocker(t)
	ctx := context.Background()

	req := &CreateRequest{Name: "test-refresh"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	refreshed, err := dm.refreshPorts(ctx, p)
	if err != nil {
		t.Fatalf("refreshPorts failed: %v", err)
	}
	if refreshed.SSHPort == 0 {
		t.Error("expected non-zero SSHPort after refresh")
	}
	if refreshed.WebPort == 0 {
		t.Error("expected non-zero WebPort after refresh")
	}
}

func TestDockerManager_CreateProject_WithHosts(t *testing.T) {
	ctx := context.Background()

	cfgWithHosts := &Config{
		ListenAddr:   ":8080",
		DefaultImage: testImage,
		Hosts: map[string]string{
			"myapp.local": "192.168.1.100",
			"db.local":    "10.0.0.5",
		},
	}
	dm, err := NewDockerManager(cfgWithHosts)
	if err != nil {
		t.Skipf("docker client unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = dm.Close()
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := dm.client.Ping(pingCtx, dockerclient.PingOptions{}); err != nil {
		t.Skipf("docker daemon unreachable: %v", err)
	}

	req := &CreateRequest{Name: "test-hosts"}
	p, err := dm.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}
	defer cleanupTestProject(t, dm, p.ID)

	// Inspect container to verify ExtraHosts are present.
	inspectResult, err := dm.client.ContainerInspect(ctx, ContainerName(p.ID), dockerclient.ContainerInspectOptions{})
	if err != nil {
		t.Fatalf("container inspect failed: %v", err)
	}
	inspect := inspectResult.Container

	expectedHosts := []string{
		"host.docker.internal:host-gateway",
		"myapp.local:192.168.1.100",
		"db.local:10.0.0.5",
	}

	for _, expected := range expectedHosts {
		found := false
		for _, h := range inspect.HostConfig.ExtraHosts {
			if h == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ExtraHost %q not found in container ExtraHosts %v", expected, inspect.HostConfig.ExtraHosts)
		}
	}
}

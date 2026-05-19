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
	// Remove volume
	if _, err := dm.client.VolumeRemove(ctx, VolumeName(id), dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
		t.Errorf("cleanup volume remove failed: %v", err)
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
	if p.Volume != VolumeName(p.ID) {
		t.Errorf("expected volume %s, got %s", VolumeName(p.ID), p.Volume)
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

	// Verify volume exists
	volResult, err := dm.client.VolumeInspect(ctx, p.Volume, dockerclient.VolumeInspectOptions{})
	if err != nil {
		t.Fatalf("volume inspect failed: %v", err)
	}
	vol := volResult.Volume
	if vol.Name != p.Volume {
		t.Errorf("expected volume name %s, got %s", p.Volume, vol.Name)
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

	_, err = dm.client.VolumeInspect(ctx, VolumeName(id), dockerclient.VolumeInspectOptions{})
	if err == nil {
		t.Error("expected volume to be removed")
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

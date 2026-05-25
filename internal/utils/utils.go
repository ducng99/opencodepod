package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"opencodepod/internal/config"
	"opencodepod/internal/docker"
	"opencodepod/internal/project"

	"github.com/containerd/errdefs"
	dockerclient "github.com/moby/moby/client"
)

const TestImage = "nginx:alpine"

func RequireDocker(t *testing.T) *docker.DockerManager {
	t.Helper()
	config.Cfg = &config.Config{
		ListenAddr:   ":8080",
		DefaultImage: TestImage,
	}
	dm, err := docker.NewDockerManager()
	if err != nil {
		t.Fatalf("docker client unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = dm.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := dm.Client.Ping(ctx, dockerclient.PingOptions{}); err != nil {
		t.Fatalf("docker daemon unreachable: %v", err)
	}

	if _, err := dm.Client.ImageInspect(ctx, TestImage); err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			pullCtx, pullCancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer pullCancel()
			pr, err := dm.Client.ImagePull(pullCtx, TestImage, dockerclient.ImagePullOptions{})
			if err != nil {
				t.Fatalf("unable to pull test image %s: %v", TestImage, err)
			}
			_, _ = io.Copy(io.Discard, pr)
			_ = pr.Close()
		} else {
			t.Fatalf("unable to inspect test image %s: %v", TestImage, err)
		}
	}

	return dm
}

func CleanupTestProject(t *testing.T, dm *docker.DockerManager, id string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	f := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", project.LabelProjectID, id))
	result, err := dm.Client.ContainerList(ctx, dockerclient.ContainerListOptions{All: true, Filters: f})
	if err != nil {
		t.Errorf("cleanup container list failed: %v", err)
		return
	}
	for _, c := range result.Items {
		if _, err := dm.Client.ContainerRemove(ctx, c.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
			t.Errorf("cleanup container remove failed: %v", err)
		}
	}
	for _, vol := range project.ProjectVolumes(id) {
		if _, err := dm.Client.VolumeRemove(ctx, vol, dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
			t.Errorf("cleanup volume remove failed: %v", err)
		}
	}
}

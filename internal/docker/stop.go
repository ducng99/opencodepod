package docker

import (
	"context"
	"fmt"

	"opencodepod/internal/project"

	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) StopProject(ctx context.Context, id string) (*project.Project, error) {
	p, err := dm.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status != "running" {
		return p, nil
	}

	if _, err := dm.Client.ContainerStop(ctx, project.ContainerName(id), dockerclient.ContainerStopOptions{}); err != nil {
		return nil, fmt.Errorf("stop: %w", err)
	}
	return dm.GetProject(ctx, id)
}

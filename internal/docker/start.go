package docker

import (
	"context"
	"fmt"

	"opencodepod/internal/project"

	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) StartProject(ctx context.Context, id string) (*project.Project, error) {
	p, err := dm.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Status == "running" {
		return p, nil
	}

	if _, err := dm.Client.ContainerStart(ctx, project.ContainerName(id), dockerclient.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}
	return dm.GetProject(ctx, id)
}

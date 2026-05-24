package docker

import (
	"context"

	"opencodepod/internal/project"
)

func (dm *DockerManager) GetProject(ctx context.Context, id string) (*project.Project, error) {
	inspect, err := dm.inspectProject(ctx, id)
	if err != nil {
		return nil, err
	}

	p := project.ProjectFromLabels(id, inspect.Config.Labels)
	populateProjectFromInspect(p, &inspect)
	return p, nil
}

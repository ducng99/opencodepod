package docker

import (
	"context"
	"fmt"

	"opencodepod/internal/project"

	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) ListProjects(ctx context.Context) ([]*project.Project, error) {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=true", project.LabelManaged))

	result, err := dm.Client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}

	projects := make([]*project.Project, 0, len(result.Items))
	for _, c := range result.Items {
		p := dm.containerToProject(&c)
		projects = append(projects, p)
	}
	return projects, nil
}

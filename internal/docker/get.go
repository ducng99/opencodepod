package docker

import (
	"context"
	"errors"
	"fmt"

	"opencodepod/internal/project"

	"github.com/containerd/errdefs"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) GetProject(ctx context.Context, id string) (*project.Project, error) {
	inspectResult, err := dm.Client.ContainerInspect(ctx, project.ContainerName(id), dockerclient.ContainerInspectOptions{})
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			return nil, fmt.Errorf("project not found: %s", id)
		}
		return nil, err
	}
	inspect := inspectResult.Container

	p := project.ProjectFromLabels(id, inspect.Config.Labels)

	healthStatus := ""
	if inspect.State.Health != nil {
		healthStatus = string(inspect.State.Health.Status)
	}
	p.Status = computeStatus(string(inspect.State.Status), healthStatus)
	p.SSHPort = hostPortFromInspect(&inspect, "22/tcp")
	p.WebPort = hostPortFromInspect(&inspect, "8080/tcp")
	return p, nil
}

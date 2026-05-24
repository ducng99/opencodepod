package docker

import (
	"context"
	"fmt"

	"opencodepod/internal/project"

	"github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) RenameProject(ctx context.Context, id string, req *project.UpdateRequest) (*project.Project, error) {
	inspect, err := dm.inspectProject(ctx, id)
	if err != nil {
		return nil, err
	}

	labels := make(map[string]string, len(inspect.Config.Labels))
	for k, v := range inspect.Config.Labels {
		labels[k] = v
	}
	labels[project.LabelName] = req.Name

	if err := dm.stopAndRemoveContainer(ctx, inspect.ID); err != nil {
		return nil, err
	}

	containerConfig := &container.Config{
		Image:        inspect.Config.Image,
		Labels:       labels,
		ExposedPorts: inspect.Config.ExposedPorts,
		Env:          inspect.Config.Env,
	}

	hostConfig := &container.HostConfig{
		PortBindings:  inspect.HostConfig.PortBindings,
		Binds:         inspect.HostConfig.Binds,
		ExtraHosts:    inspect.HostConfig.ExtraHosts,
		RestartPolicy: inspect.HostConfig.RestartPolicy,
	}

	createResult, err := dm.Client.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config:     containerConfig,
		HostConfig: hostConfig,
		Name:       project.ContainerName(id),
	})
	if err != nil {
		return nil, fmt.Errorf("container create: %w", err)
	}

	if err := dm.injectSecrets(ctx, createResult.ID); err != nil {
		return nil, err
	}

	if err := dm.startContainer(ctx, createResult.ID); err != nil {
		return nil, err
	}

	inspectResult, err := dm.Client.ContainerInspect(ctx, createResult.ID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("container inspect: %w", err)
	}
	inspect = inspectResult.Container

	p := project.ProjectFromLabels(id, inspect.Config.Labels)
	p.Image = inspect.Config.Image
	p.Volumes = project.ProjectVolumes(id)

	populateProjectFromInspect(p, &inspect)

	return p, nil
}

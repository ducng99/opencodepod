package docker

import (
	"context"
	"fmt"
	"io"

	"opencodepod/internal/project"

	"github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) UpgradeProject(ctx context.Context, id string) (*project.Project, error) {
	inspect, err := dm.inspectProject(ctx, id)
	if err != nil {
		return nil, err
	}
	image := inspect.Config.Image

	// Pull the latest image. The daemon streams JSON progress messages; we must
	// drain the entire body so the HTTP connection completes and the pull finishes.
	pr, err := dm.Client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if _, inspectErr := dm.Client.ImageInspect(ctx, image); inspectErr != nil {
			return nil, fmt.Errorf("image pull failed and no local image found: %w", err)
		}
	}

	// Always recreate the container so config changes (e.g., mounts) are applied.
	// The volume is preserved.
	if err := dm.stopAndRemoveContainer(ctx, inspect.ID); err != nil {
		return nil, err
	}

	p := project.ProjectFromLabels(id, inspect.Config.Labels)

	containerConfig := &container.Config{
		Image:        image,
		Labels:       inspect.Config.Labels,
		ExposedPorts: inspect.Config.ExposedPorts,
		Env:          dm.buildEnv(p),
	}

	hostConfig := &container.HostConfig{
		PortBindings:  inspect.HostConfig.PortBindings,
		Binds:         dm.buildBinds(id, p.ContainerUser),
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

	if err := dm.injectSecrets(ctx, createResult.ID, p.ContainerUser, p.Stacks); err != nil {
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

	p = project.ProjectFromLabels(id, inspect.Config.Labels)
	p.Image = image
	p.Volumes = project.ProjectVolumes(id)

	populateProjectFromInspect(p, &inspect)

	return p, nil
}

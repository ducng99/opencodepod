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

	// Capture the container's actual image ID so we can tell whether the pull
	// actually changed anything. We must use inspect.Image (the ID the container
	// was created from) rather than ImageInspect by tag, because the local tag
	// may already point to a newer image while the container still runs the old one.
	oldImageID := inspect.Image

	// Pull the latest image. The daemon streams JSON progress messages; we must
	// drain the entire body so the HTTP connection completes and the pull finishes.
	pr, err := dm.Client.ImagePull(ctx, image, dockerclient.ImagePullOptions{})
	if err == nil {
		_, _ = io.Copy(io.Discard, pr)
		_ = pr.Close()
	} else {
		if oldImageID == "" {
			return nil, fmt.Errorf("image pull failed and no local image found: %w", err)
		}
		return nil, fmt.Errorf("image pull failed; container was not upgraded because no newer image could be retrieved: %w", err)
	}

	// Check whether the image actually changed.
	newInspect, err := dm.Client.ImageInspect(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("image inspect after pull: %w", err)
	}
	if oldImageID != "" && oldImageID == newInspect.ID {
		// Same image — no need to recreate the container.
		return dm.GetProject(ctx, id)
	}

	// Newer image obtained. Stop and remove the old container (volume is preserved).
	if err := dm.stopAndRemoveContainer(ctx, inspect.ID); err != nil {
		return nil, err
	}

	containerConfig := &container.Config{
		Image:        image,
		Labels:       inspect.Config.Labels,
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
	p.Image = image
	p.Volumes = project.ProjectVolumes(id)

	populateProjectFromInspect(p, &inspect)

	return p, nil
}

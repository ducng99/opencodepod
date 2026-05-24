package docker

import (
	"context"
	"errors"
	"fmt"

	"opencodepod/internal/project"

	"github.com/containerd/errdefs"
	dockerclient "github.com/moby/moby/client"
)

func (dm *DockerManager) DeleteProject(ctx context.Context, id string) error {
	filter := dockerclient.Filters{}.Add("label", fmt.Sprintf("%s=%s", project.LabelProjectID, id))
	result, err := dm.Client.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
	if err != nil {
		return err
	}
	for _, c := range result.Items {
		if _, err := dm.Client.ContainerRemove(ctx, c.ID, dockerclient.ContainerRemoveOptions{Force: true}); err != nil {
			return fmt.Errorf("remove container: %w", err)
		}
	}
	for _, vol := range project.ProjectVolumes(id) {
		if _, err := dm.Client.VolumeRemove(ctx, vol, dockerclient.VolumeRemoveOptions{Force: true}); err != nil {
			if !errors.Is(err, errdefs.ErrNotFound) {
				return fmt.Errorf("remove volume %s: %w", vol, err)
			}
		}
	}
	return nil
}

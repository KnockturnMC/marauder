package servermgr

import (
	"context"
	"fmt"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/docker/docker/api/types"

	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types/container"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

func (d DockerBasedManager) Stop(ctx context.Context, serverModel networkmodel.ServerModel) error {
	serverName := d.computeUniqueDockerContainerNameFor(serverModel)

	timeoutInSeconds := int((time.Minute * 5).Seconds())
	if err := d.DockerClient.ContainerStop(ctx, serverName, container.StopOptions{
		Timeout: &timeoutInSeconds,
	}); err != nil {
		if utils.CheckDockerError(err, client.IsErrNotFound) {
			return nil // Server is just not running
		}

		return fmt.Errorf("failed to stop container %s: %w", serverName, err)
	}

	if err := d.DockerClient.ContainerRemove(ctx, serverName, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

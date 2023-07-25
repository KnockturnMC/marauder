package servermgr

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

func (d DockerBasedManager) Stop(ctx context.Context, serverModel networkmodel.ServerModel) error {
	serverName := d.computeUniqueDockerContainerNameFor(serverModel)

	if err := d.DockerClient.ContainerStop(ctx, serverName, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", serverName, err)
	}

	return nil
}

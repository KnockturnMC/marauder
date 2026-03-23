package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
	"github.com/sirupsen/logrus"
)

// ErrContainerNotRemovedInTime is returned if the stop logic did not find the container to be removed in time.
var ErrContainerNotRemovedInTime = errors.New("not removed in time")

func (d DockerBasedManager) Stop(ctx context.Context, serverModel networkmodel.ServerModel) error {
	serverName := d.computeUniqueDockerContainerNameFor(serverModel)

	timeout := time.Minute * 5
	deadline := time.Now().Add(timeout)
	withDeadline, cancelFunction := context.WithDeadline(ctx, deadline)
	defer cancelFunction()

	// If the server has a management socket path assigned, first try to send a server stop note.
	if serverModel.ManagementSocketPath != "" {
		attemptShutdownViaManagementSocket(withDeadline, d, serverModel, serverName)
	}

	if err := d.DockerClient.ContainerStop(ctx, serverName, container.StopOptions{
		Timeout: new(int(timeout.Seconds())),
	}); err != nil {
		if !utils.CheckDockerError(err, errdefs.IsNotFound) {
			return fmt.Errorf("failed to stop container %s: %w", serverName, err)
		}
	}

	if d.AutoRemoveContainers {
		err := d.awaitAutoRemoval(ctx, serverModel, deadline)
		if err != nil {
			return err
		}
	} else {
		if err := d.DockerClient.ContainerRemove(ctx, serverName, container.RemoveOptions{}); err != nil {
			if !utils.CheckDockerError(err, errdefs.IsNotFound) {
				return fmt.Errorf("failed to manually remove the container %s: %w", serverName, err)
			}
		}
	}

	return nil
}

// awaitAutoRemoval awaits a container that was constructed with auto removal to remove itself after stopping.
func (d DockerBasedManager) awaitAutoRemoval(ctx context.Context, serverModel networkmodel.ServerModel, deadline time.Time) error {
	// Await removal via AutoRemove: true flag in container creation.
	for !time.Now().After(deadline) {
		if _, err := d.retrieveContainerInfo(ctx, serverModel); err != nil {
			if utils.CheckDockerError(err, errdefs.IsNotFound) {
				return nil
			}

			return fmt.Errorf("failed to retrieve container info: %w", err)
		}
	}
	return ErrContainerNotRemovedInTime
}

func attemptShutdownViaManagementSocket(
	ctx context.Context,
	dockerBasedManager DockerBasedManager,
	serverModel networkmodel.ServerModel,
	serverName string,
) {
	err := dockerBasedManager.ExchangeManagementMessage(
		ctx, serverModel,
		marauderpb.ServerShutdownRequest_builder{Reason: new("Shutdown via marauder")}.Build(),
		marauderpb.ServerShutdownRequest_Response_builder{}.Build(),
	)
	if err != nil {
		logrus.Warn("failed to stop server via management message ", err)
	} else {
		waitChan, errChan := dockerBasedManager.DockerClient.ContainerWait(ctx, serverName, container.WaitConditionNotRunning)
	reader:
		for {
			select {
			case <-waitChan:
				break reader
			case err = <-errChan:
				if !utils.CheckDockerError(err, errdefs.IsNotFound) {
					logrus.Warn("failed to await stopping container via management message ", err)
				}
				break reader
			}
		}
	}
}

package servermgr

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

func (d DockerBasedManager) Start(ctx context.Context, server networkmodel.ServerModel) error {
	_, err := d.retrieveContainerInfo(ctx, server)
	if err == nil {
		return nil // It is running
	}
	if err != nil && !utils.CheckDockerError(err, client.IsErrNotFound) {
		return fmt.Errorf("failed to retrieve the container information: %w", err)
	}

	reader, err := d.DockerClient.ImagePull(ctx, server.Image, types.ImagePullOptions{
		RegistryAuth: d.DockerEncodedAuth,
	})
	if err != nil {
		return fmt.Errorf("failed to pull image for server start: %w", err)
	}

	if _, err := io.ReadAll(reader); err != nil {
		return fmt.Errorf("failed to consume image pull reader: %w", err)
	}

	location, err := d.computeServerFolderLocation(server)
	if err != nil {
		return fmt.Errorf("failed to compute server folder location: %w", err)
	}

	if err := os.MkdirAll(location, 0o700); err != nil {
		return fmt.Errorf("failed to mkdir the server data folder %s: %w", location, err)
	}

	err = d.starDockerContainer(ctx, server, location)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerBasedManager) starDockerContainer(ctx context.Context, server networkmodel.ServerModel, systemPath string) error {
	computeServerFolderLocation, err := d.DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:        server.Image,
			ExposedPorts: map[nat.Port]struct{}{nat.Port(strconv.Itoa(server.Port)): {}},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{{
				Type:   mount.TypeBind,
				Source: systemPath,
				Target: "/home/server",
			}},
		},
		nil,
		nil,
		d.computeUniqueDockerContainerNameFor(server),
	)
	if err != nil {
		return fmt.Errorf("failed to create server container: %w", err)
	}

	for _, networkConfiguration := range server.Networks {
		if err := d.DockerClient.NetworkConnect(ctx, networkConfiguration.NetworkName, computeServerFolderLocation.ID, &network.EndpointSettings{
			IPAddress: networkConfiguration.IPV4Address,
		}); err != nil {
			_ = d.DockerClient.ContainerRemove(ctx, computeServerFolderLocation.ID, types.ContainerRemoveOptions{})
			return fmt.Errorf("failed to connect server to network %s: %w", networkConfiguration.NetworkName, err)
		}
	}

	if err := d.DockerClient.ContainerStart(ctx, computeServerFolderLocation.ID, types.ContainerStartOptions{}); err != nil {
		_ = d.DockerClient.ContainerRemove(ctx, computeServerFolderLocation.ID, types.ContainerRemoveOptions{})
		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

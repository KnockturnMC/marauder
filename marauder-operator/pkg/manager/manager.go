package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"

	dockerClient "github.com/docker/docker/client"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/fileeq"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// ErrNoMatchingDiskConfig is returned for a deployment that is not matched by any disk matcher.
var ErrNoMatchingDiskConfig = errors.New("no matching disk config")

type Manager interface {
	// Stop shuts don the server passed to the manager.
	Stop(ctx context.Context, server networkmodel.ServerModel) error

	// Start starts the server model. If the given server is running, this method is a NOOP.
	Start(ctx context.Context, server networkmodel.ServerModel) error

	// UpdateDeployments updates all deployments currently defined on the server.
	UpdateDeployments(
		ctx context.Context,
		server networkmodel.ServerModel,
		requiresRestart bool,
		failOnUnexpectedOldFilesOnDisk bool,
	) error
}

// FolderOwner defines the folder owner for the docker based mounting.
type FolderOwner struct {
	GID int `yaml:"gid"`
	UID int `yaml:"uid"`
}

// EnvironmentDiskConfig describes a specifics environment's disk configuration.
type EnvironmentDiskConfig struct {
	ServerDataPathTemplate string       `yaml:"serverDataPathTemplate"`
	FolderOwner            *FolderOwner `yaml:"folderOwner,omitempty"`
}

type DiskPathMapping map[string]EnvironmentDiskConfig

// The DockerBasedManager implements the manager interface and manages server via a docker client.
type DockerBasedManager struct {
	DockerClient      *dockerClient.Client
	DockerEncodedAuth string

	ControllerClient      controller.DownloadingClient
	AutoRemoveContainers  bool
	ContainerMemoryBuffer int64

	DiskPathMapping DiskPathMapping

	FileEqualityRegistry fileeq.FileEqualityRegistry
}

// FindDiskConfig locates the disk config for a specific environment.
func (d DockerBasedManager) FindDiskConfig(server networkmodel.ServerModel) (EnvironmentDiskConfig, error) {
	pathsToQuery := []string{
		fmt.Sprintf("%s/%s", server.Environment, server.Name),
		server.Environment,
		"*",
	}

	for _, path := range pathsToQuery {
		mapping, found := d.DiskPathMapping[path]
		if found {
			return mapping, nil
		}
	}

	return EnvironmentDiskConfig{}, fmt.Errorf(
		"failed to find config for search path [%s]: %w",
		strings.Join(pathsToQuery, ", "),
		ErrNoMatchingDiskConfig,
	)
}

// computeUniqueDockerContainerNameFor computes the unique docker container name for a given server managed by
// the operator.
func (d DockerBasedManager) computeUniqueDockerContainerNameFor(server networkmodel.ServerModel) string {
	visibility := "local"
	if len(server.HostPorts) > 0 {
		visibility = "public"
	}

	return fmt.Sprintf("%s-minecraft-%s-%s", visibility, server.Environment, server.Name)
}

// computeServerFolderLocation computes the on-disk server folder location on the host.
// The operator can extract artefacts into said directory when instructed to.
func (d DockerBasedManager) computeServerFolderLocation(server networkmodel.ServerModel) (string, error) {
	diskConfig, err := d.FindDiskConfig(server)
	if err != nil {
		return "", fmt.Errorf("failed to find disk config: %w", err)
	}

	toString, err := utils.ExecuteStringTemplateToString(diskConfig.ServerDataPathTemplate, server)
	if err != nil {
		return "", fmt.Errorf("failed to expand string template: %w", err)
	}

	return toString, nil
}

// retrieveContainerInfo retrieves the container json about a specific server models potential container.
//
//nolint:unparam
func (d DockerBasedManager) retrieveContainerInfo(ctx context.Context, server networkmodel.ServerModel) (container.InspectResponse, error) {
	identifier := d.computeUniqueDockerContainerNameFor(server)
	inspectResponse, err := d.DockerClient.ContainerInspect(ctx, identifier)
	if err != nil {
		return container.InspectResponse{}, fmt.Errorf("failed to inspect container %s: %w", identifier, err)
	}

	return inspectResponse, nil
}

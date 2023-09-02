package manager

import (
	"context"
	"fmt"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg/controller"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
)

type Manager interface {
	// Stop shuts don the server passed to the manager.
	Stop(ctx context.Context, server networkmodel.ServerModel) error

	// Start starts the server model. If the given server is running, this method is a NOOP.
	Start(ctx context.Context, server networkmodel.ServerModel) error

	// UpdateDeployments updates all deployments currently defined on the server.
	UpdateDeployments(ctx context.Context, server networkmodel.ServerModel) error
}

// The DockerBasedManager implements the manager interface and manages server via a docker client.
type DockerBasedManager struct {
	DockerClient           *dockerClient.Client
	DockerEncodedAuth      string
	ControllerClient       controller.DownloadingClient
	AutoRemoveContainers   bool
	ServerDataPathTemplate string
}

// computeUniqueDockerContainerNameFor computes the unique docker container name for a given server managed by
// the operator.
func (d DockerBasedManager) computeUniqueDockerContainerNameFor(server networkmodel.ServerModel) string {
	return fmt.Sprintf("marauder-server-%s-%s-%s", server.Environment, server.Name, strings.ReplaceAll(server.UUID.String(), "-", ""))
}

// computeServerFolderLocation computes the on-disk server folder location on the host.
// The operator can extract artefacts into said directory when instructed to.
func (d DockerBasedManager) computeServerFolderLocation(server networkmodel.ServerModel) (string, error) {
	toString, err := utils.ExecuteStringTemplateToString(d.ServerDataPathTemplate, server)
	if err != nil {
		return "", fmt.Errorf("failed to expand string template: %w", err)
	}

	return toString, nil
}

// retrieveContainerInfo retrieves the container json about a specific server models potential container.
//
//nolint:unparam
func (d DockerBasedManager) retrieveContainerInfo(ctx context.Context, server networkmodel.ServerModel) (types.ContainerJSON, error) {
	identifier := d.computeUniqueDockerContainerNameFor(server)
	container, err := d.DockerClient.ContainerInspect(ctx, identifier)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("failed to inspect container %s: %w", identifier, err)
	}

	return container, nil
}

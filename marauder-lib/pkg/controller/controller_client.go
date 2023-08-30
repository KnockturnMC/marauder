package controller

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/samber/mo"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/google/uuid"
)

// ErrIncorrectReferenceFormat is returned if the reference is in a wrong format.
var ErrIncorrectReferenceFormat = errors.New("incorrect format")

// The Client is responsible for interacting with the controller from the operator side.
type Client interface {
	// ResolveArtefactReference resolves a reference to a specific artefact to its uuid.
	ResolveArtefactReference(ctx context.Context, reference string) (uuid.UUID, error)

	// ResolveServerReference resolves a reference to a specific server to its uuid.
	ResolveServerReference(ctx context.Context, reference string) (uuid.UUID, error)

	// FetchArtefact fetches an artefact model from the controller given the uuid.
	FetchArtefact(ctx context.Context, artefact uuid.UUID) (networkmodel.ArtefactModel, error)

	// FetchArtefactByIdentifierAndVersion fetches an artefact model from the controller given the identifier and version.
	FetchArtefactByIdentifierAndVersion(ctx context.Context, identifier, version string) (networkmodel.ArtefactModel, error)

	// FetchServer fetches a server model from the controller given the uuid.
	FetchServer(ctx context.Context, server uuid.UUID) (networkmodel.ServerModel, error)

	// FetchArtefacts fetches artefact models from the controller given the identifier.
	FetchArtefacts(ctx context.Context, identifier string) ([]networkmodel.ArtefactModel, error)

	// FetchServers fetches server models from the controller given their environment.
	FetchServers(ctx context.Context, environment string) ([]networkmodel.ServerModel, error)

	// FetchServerStateArtefacts fetches the artefacts defined for the specific state on the given server.
	FetchServerStateArtefacts(ctx context.Context, server uuid.UUID, state networkmodel.ServerStateType) ([]networkmodel.ArtefactModel, error)

	// FetchUpdatesFor fetches all outstanding updates for a server by its uuid.
	FetchUpdatesFor(ctx context.Context, server uuid.UUID) ([]networkmodel.ArtefactVersionMissmatch, error)

	// FetchManifest fetches a manifest based on its uuid.
	FetchManifest(ctx context.Context, artefact uuid.UUID) (filemodel.Manifest, error)

	// ExecuteActionOn posts a lifecycle action to the operator of the server for the given server.
	ExecuteActionOn(ctx context.Context, server uuid.UUID, action networkmodel.LifecycleChangeActionType) error

	// PublishArtefact publishes the artefact read from the given readers to the controller.
	// The method returns the status code of the response for further usage.
	PublishArtefact(ctx context.Context, artefact, signature io.Reader) (networkmodel.ArtefactModel, mo.Option[int], error)

	// UpdateState attempts to update the controller about a servers new state for the specific artefact.
	UpdateState(ctx context.Context, server uuid.UUID, state networkmodel.ServerStateType, artefactIdentifier string, artefact uuid.UUID) error
}

// HTTPClient implements the Client interface by using the controllers rest API.
type HTTPClient struct {
	*http.Client
	ControllerURL string
}

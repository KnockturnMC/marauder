package controller

import (
	"context"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/lib/pkg/worker"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/google/uuid"
)

// The Client is responsible for interacting with the controller from the operator side.
type Client interface {
	// FetchServer fetches a server model from the controller given the uuid.
	FetchServer(ctx context.Context, server uuid.UUID) (networkmodel.ServerModel, error)

	// FetchUpdatesFor fetches all outstanding updates for a server by its uuid.
	FetchUpdatesFor(ctx context.Context, server uuid.UUID) ([]networkmodel.VersionDiff, error)

	// FetchManifest fetches a manifest based on its uuid.
	FetchManifest(ctx context.Context, artefact uuid.UUID) (filemodel.Manifest, error)

	// UpdateIsState attempts to update the controller about a servers new is state for the specific artefact.
	UpdateIsState(ctx context.Context, server uuid.UUID, artefactIdentifier string, artefact uuid.UUID) error

	// DownloadArtefact downloads the artefact specified with the given uuid a local cache folder and
	// returns the full path to the downloaded file.
	DownloadArtefact(ctx context.Context, artefactUUID uuid.UUID) (string, error)
}

// HTTPClient implements the Client interface by using the controllers rest API.
type HTTPClient struct {
	*http.Client
	ControllerURL   string
	DownloadService worker.DownloadService
}

func (h *HTTPClient) FetchServer(ctx context.Context, server uuid.UUID) (networkmodel.ServerModel, error) {
	model, err := utils.HTTPGetAndBind(ctx, h.Client, h.ControllerURL+"/server/"+server.String(), networkmodel.ServerModel{})
	if err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed http get: %w", err)
	}

	return model, nil
}

func (h *HTTPClient) FetchUpdatesFor(ctx context.Context, server uuid.UUID) ([]networkmodel.VersionDiff, error) {
	diffs, err := utils.HTTPGetAndBind(ctx, h.Client, h.ControllerURL+"/deployment/"+server.String()+"/update", make([]networkmodel.VersionDiff, 0))
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return diffs, nil
}

// FetchManifest fetches a manifest based on its uuid.
func (h *HTTPClient) FetchManifest(ctx context.Context, artefact uuid.UUID) (filemodel.Manifest, error) {
	manifest, err := utils.HTTPGetAndBind(ctx, h.Client, h.ControllerURL+"/artefact/"+artefact.String()+"/download/manifest", filemodel.Manifest{})
	if err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed http get: %w", err)
	}

	return manifest, nil
}

func (h *HTTPClient) DownloadArtefact(ctx context.Context, artefactUUID uuid.UUID) (string, error) {
	downloadedFile, err := h.DownloadService.Download(
		ctx,
		fmt.Sprintf("%s/artefact/%s/download", h.ControllerURL, artefactUUID.String()),
		fmt.Sprintf("%s.tar.gz", artefactUUID.String()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to download artefact: %w", err)
	}

	return downloadedFile, nil
}

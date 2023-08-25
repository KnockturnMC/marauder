package controller

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/worker"
	"github.com/google/uuid"
)

// DownloadingClient represents a controller client that can also download large files from the controller.
type DownloadingClient interface {
	Client

	// DownloadArtefact downloads the artefact specified with the given uuid a local cache folder and
	// returns the full path to the downloaded file.
	DownloadArtefact(ctx context.Context, artefactUUID uuid.UUID) (string, error)
}

// DownloadingHTTPClient is a http based implementation of the DownloadService interface.
type DownloadingHTTPClient struct {
	HTTPClient

	DownloadService worker.DownloadService
}

func (h *DownloadingHTTPClient) DownloadArtefact(ctx context.Context, artefactUUID uuid.UUID) (string, error) {
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

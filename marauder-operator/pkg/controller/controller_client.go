package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
}

// HTTPClient implements the Client interface by using the controllers rest API.
type HTTPClient struct {
	http.Client
	ControllerURL string
}

// getAndBind performs a get request using the http client at the given path and binds the result into
// the passed struct.
// If a response code that is not 200<=code<=400, an error is returned.
func getAndBind[T any](ctx context.Context, client *HTTPClient, path string, bindTarget T) (T, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, client.ControllerURL+path, &bytes.Buffer{})
	if err != nil {
		return bindTarget, fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := client.Do(request)
	if err != nil {
		return bindTarget, fmt.Errorf("failed to perform get request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return bindTarget, fmt.Errorf("failed to read body of get request: %w", err)
	}

	if !worker.IsOkayStatusCode(resp.StatusCode) {
		return bindTarget, fmt.Errorf("controller returned '%s' (%d): %w", string(body), resp.StatusCode, worker.ErrBadStatusCode)
	}

	if err := json.Unmarshal(body, &bindTarget); err != nil {
		return bindTarget, fmt.Errorf("failed to bind response to bind target: %w", err)
	}

	return bindTarget, nil
}

func (h *HTTPClient) FetchServer(ctx context.Context, server uuid.UUID) (networkmodel.ServerModel, error) {
	model, err := getAndBind(ctx, h, "/server/"+server.String(), networkmodel.ServerModel{})
	if err != nil {
		return networkmodel.ServerModel{}, err
	}

	return model, nil
}

func (h *HTTPClient) FetchUpdatesFor(ctx context.Context, server uuid.UUID) ([]networkmodel.VersionDiff, error) {
	diffs, err := getAndBind(ctx, h, "/deployment/"+server.String()+"/update", make([]networkmodel.VersionDiff, 0))
	if err != nil {
		return nil, err
	}

	return diffs, nil
}

// FetchManifest fetches a manifest based on its uuid.
func (h *HTTPClient) FetchManifest(ctx context.Context, artefact uuid.UUID) (filemodel.Manifest, error) {
	manifest, err := getAndBind(ctx, h, "/artefact/"+artefact.String()+"/download/manifest", filemodel.Manifest{})
	if err != nil {
		return filemodel.Manifest{}, err
	}

	return manifest, nil
}

func (h *HTTPClient) UpdateIsState(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}

package controller

import (
	"context"
	"fmt"
	"strconv"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/google/uuid"
)

func (h *HTTPClient) FetchServer(ctx context.Context, server uuid.UUID) (networkmodel.ServerModel, error) {
	model, err := utils.HTTPGetAndBind(ctx, h.Client, h.ControllerURL+"/server/"+server.String(), networkmodel.ServerModel{})
	if err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed http get: %w", err)
	}

	return model, nil
}

func (h *HTTPClient) FetchMissmatchesFor(
	ctx context.Context,
	server uuid.UUID,
	requiresRestart bool,
) ([]networkmodel.ArtefactVersionMissmatch, error) {
	diffs, err := utils.HTTPGetAndBind(
		ctx,
		h.Client,
		fmt.Sprintf("%s/server/%s/state/update?requiresRestart=%s", h.ControllerURL, server.String(), strconv.FormatBool(requiresRestart)),
		make([]networkmodel.ArtefactVersionMissmatch, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return diffs, nil
}

// FetchArtefact fetches an artefact model from the controller given the uuid.
func (h *HTTPClient) FetchArtefact(ctx context.Context, artefact uuid.UUID) (networkmodel.ArtefactModel, error) {
	bind, err := utils.HTTPGetAndBind(ctx, h.Client, fmt.Sprintf("%s/artefact/%s", h.ControllerURL, artefact.String()), networkmodel.ArtefactModel{})
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed http get: %w", err)
	}

	return bind, nil
}

// FetchArtefactByIdentifierAndVersion fetches an artefact model from the controller given the identifier and version.
func (h *HTTPClient) FetchArtefactByIdentifierAndVersion(ctx context.Context, identifier, version string) (networkmodel.ArtefactModel, error) {
	bind, err := utils.HTTPGetAndBind(ctx, h.Client, fmt.Sprintf(
		"%s/artefacts/%s/%s",
		h.ControllerURL,
		identifier,
		version,
	), networkmodel.ArtefactModel{})
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed http get: %w", err)
	}

	return bind, nil
}

// FetchArtefacts fetches artefact models from the controller given the identifier.
func (h *HTTPClient) FetchArtefacts(ctx context.Context, identifier string) ([]networkmodel.ArtefactModel, error) {
	bind, err := utils.HTTPGetAndBind(
		ctx,
		h.Client,
		fmt.Sprintf("%s/artefacts/%s", h.ControllerURL, identifier),
		make([]networkmodel.ArtefactModel, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return bind, nil
}

// FetchServers fetches server models from the controller given their environment.
func (h *HTTPClient) FetchServers(ctx context.Context, environment string) ([]networkmodel.ServerModel, error) {
	bind, err := utils.HTTPGetAndBind(
		ctx,
		h.Client,
		fmt.Sprintf("%s/servers/%s", h.ControllerURL, environment),
		make([]networkmodel.ServerModel, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return bind, nil
}

// FetchServerStateArtefacts fetches the artefacts defined for the specific state on the given server.
func (h *HTTPClient) FetchServerStateArtefacts(
	ctx context.Context,
	server uuid.UUID,
	state networkmodel.ServerStateType,
) ([]networkmodel.ArtefactModel, error) {
	artefacts, err := utils.HTTPGetAndBind(
		ctx,
		h.Client,
		fmt.Sprintf("%s/server/%s/state/%s", h.ControllerURL, server, state),
		make([]networkmodel.ArtefactModel, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return artefacts, nil
}

// FetchManifest fetches a manifest based on its uuid.
func (h *HTTPClient) FetchManifest(ctx context.Context, artefact uuid.UUID) (filemodel.Manifest, error) {
	manifest, err := utils.HTTPGetAndBind(ctx, h.Client, h.ControllerURL+"/artefact/"+artefact.String()+"/download/manifest", filemodel.Manifest{
		Files:            make(filemodel.FileReferenceCollection, 0),
		BuildInformation: &filemodel.BuildInformation{},
	})
	if err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed http get: %w", err)
	}

	return manifest, nil
}

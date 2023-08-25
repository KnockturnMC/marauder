package controller

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/google/uuid"
)

// PostToOperator posts a lifecycle action to the operator of the server for the given server
func (h *HTTPClient) PostToOperator(ctx context.Context, server uuid.UUID, action networkmodel.LifecycleChangeActionType) error {
	bind, err := utils.HTTPGetAndBind(ctx, h.Client, fmt.Sprintf("%s/artefact/%s", h.ControllerURL, artefact.String()), networkmodel.ArtefactModel{})
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed http get: %w", err)
	}

	return bind, nil
}

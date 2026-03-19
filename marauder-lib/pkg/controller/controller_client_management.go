package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
)

// ManageServerPlayers fetches the players currently on the server.
func (h *HTTPClient) ManageServerPlayers(ctx context.Context, server uuid.UUID) ([]marauderpb.Player, error) {
	artefacts, err := utils.HTTPGetAndBind(
		ctx,
		h.Client,
		fmt.Sprintf("%s/operator/%s/proxy/server/%s/management/players", h.ControllerURL, server, server),
		make([]marauderpb.Player, 0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed http get: %w", err)
	}

	return artefacts, nil
}

// ManageServerToggleSave fetches all players currently on the passed server.
func (h *HTTPClient) ManageServerToggleSave(ctx context.Context, server uuid.UUID, shouldSave bool) error {
	body := networkmodel.ManagementToggleSaveBody{ShouldSave: shouldSave}
	bodyAsStr, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	resp, err := utils.PerformHTTPRequest(
		ctx,
		h.Client,
		"POST",
		fmt.Sprintf("%s/operator/%s/proxy/server/%s/management/togglesave", h.ControllerURL, server, server),
		"application/json",
		bytes.NewBuffer(bodyAsStr),
	)
	if err != nil {
		return fmt.Errorf("failed to post http: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if err := utils.IsOkayStatusCodeOrErrorWithBody(resp); err != nil {
		return fmt.Errorf("failed to execute action on: %w", err)
	}
	return nil
}

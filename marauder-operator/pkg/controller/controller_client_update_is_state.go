package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/google/uuid"
)

func (h *HTTPClient) UpdateIsState(ctx context.Context, server uuid.UUID, artefactIdentifier string, artefactUUID uuid.UUID) error {
	updateRequest := networkmodel.UpdateServerStateRequest{
		ArtefactIdentifier: artefactIdentifier,
		ArtefactUUID:       artefactUUID,
	}
	updateRequestMarshalled, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		fmt.Sprintf("%s/server/%s/state/is", h.ControllerURL, server.String()),
		bytes.NewBuffer(updateRequestMarshalled),
	)
	if err != nil {
		return fmt.Errorf("faild to create patch request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json") // Set content type as json, we are patching a json body in.

	httpResp, err := h.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute http patch request: %w", err)
	}

	defer func() { _ = httpResp.Body.Close() }()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body of get request: %w", err)
	}

	if !utils.IsOkayStatusCode(httpResp.StatusCode) {
		return fmt.Errorf("controller returned '%s' (%d): %w", string(body), httpResp.StatusCode, utils.ErrBadStatusCode)
	}

	return nil
}

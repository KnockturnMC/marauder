package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

func (h *HTTPClient) UpdateState(
	ctx context.Context,
	server uuid.UUID,
	state networkmodel.ServerStateType,
	updateRequest networkmodel.UpdateServerStateRequest,
) error {
	updateRequestMarshalled, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal update request: %w", err)
	}

	method := http.MethodPatch
	if updateRequest.ArtefactUUID == nil {
		method = http.MethodDelete
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s/server/%s/state/%s", h.ControllerURL, server.String(), state),
		bytes.NewBuffer(updateRequestMarshalled),
	)
	if err != nil {
		return fmt.Errorf("faild to create patch request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json") // Set content type as json, we are patching a json body in.

	httpResp, err := h.Do(httpReq)
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

package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/google/uuid"
)

// ExecuteActionOn posts a lifecycle action to the operator of the server for the given server.
func (h *HTTPClient) ExecuteActionOn(ctx context.Context, server uuid.UUID, action networkmodel.LifecycleChangeActionType) error {
	response, err := utils.PerformHTTPRequest(
		ctx,
		h.Client,
		http.MethodPost,
		fmt.Sprintf("%s/operator/%s/server/%s/%s", h.ControllerURL, server, server, action),
		"application/json",
		&bytes.Buffer{},
	)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if !utils.IsOkayStatusCode(response.StatusCode) {
		faultyResponseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read not-okay controller response: %w", err)
		}

		return fmt.Errorf("failed to apply action %s on %s: %s: %w", action, server, faultyResponseBody, utils.ErrStatusCodeUnrecoverable)
	}

	return nil
}

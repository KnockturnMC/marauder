package controller

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// ExecuteActionOn posts a lifecycle action to the operator of the server for the given server.
func (h *HTTPClient) ExecuteActionOn(ctx context.Context, server uuid.UUID, action networkmodel.LifecycleAction, delay time.Duration) error {
	delayQuery := ""
	if delay != 0 {
		delayQuery = "delay=" + delay.String()
	}

	response, err := utils.PerformHTTPRequest(
		ctx,
		h.Client,
		http.MethodPost,
		fmt.Sprintf("%s/operator/%s/lifecycle/%s?%s", h.ControllerURL, server, action, delayQuery),
		"application/json",
		&bytes.Buffer{},
	)
	if err != nil {
		return fmt.Errorf("failed http request: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if err := utils.IsOkayStatusCodeOrErrorWithBody(response); err != nil {
		return fmt.Errorf("failed to execute action on: %w", err)
	}

	return nil
}

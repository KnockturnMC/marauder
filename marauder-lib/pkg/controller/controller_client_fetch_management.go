package controller

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
)

// FetchServerPlayers fetches the players currently on the server.
func (h *HTTPClient) FetchServerPlayers(ctx context.Context, server uuid.UUID) ([]marauderpb.Player, error) {
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

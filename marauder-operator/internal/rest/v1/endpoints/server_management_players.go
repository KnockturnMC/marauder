package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
	"github.com/knockturnmc/marauder/marauder-operator/pkg/manager"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
)

func ServerManagementPlayers(
	serverManager manager.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		server, err := getServerModelKey(context)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, err))
			return
		}

		var manageResponse marauderpb.ServerPlayerRequest_Response
		if err := serverManager.ExchangeManagementMessage(context, server, &marauderpb.ServerPlayerRequest{}, &manageResponse); err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to request from server %s: %w", server.UUID, err),
			))

			return
		}

		mappedPlayers := make([]networkmodel.ManagementPlayer, 0, len(manageResponse.GetPlayers()))
		for _, player := range manageResponse.GetPlayers() {
			mappedPlayers = append(mappedPlayers, networkmodel.ManagementPlayer{
				UUID: player.GetUuid(),
				Name: player.GetName(),
			})
		}

		context.JSONP(http.StatusOK, networkmodel.ManagementPlayersResponse{
			LastPlayerLoginTimestamp:  manageResponse.GetLastPlayerLoginTimestamp().AsTime(),
			LastPlayerLogoutTimestamp: manageResponse.GetLastPlayerLogoutTimestamp().AsTime(),
			Players:                   mappedPlayers,
		})
	}
}

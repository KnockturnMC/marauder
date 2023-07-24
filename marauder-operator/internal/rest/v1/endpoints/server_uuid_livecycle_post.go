package endpoints

import (
	"fmt"
	"net/http"
	"strings"

	"gitea.knockturnmc.com/marauder/operator/pkg/servermgr"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"
	"gitea.knockturnmc.com/marauder/operator/pkg/controller"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ServerLifecycleActionPost(
	operatorIdentifier string,
	controllerClient controller.Client,
	serverManager servermgr.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("uuid")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		action := networkmodel.LifecycleChangeActionType(strings.ToLower(context.Param("action")))
		if !networkmodel.KnownLifecycleChangeActionType(action) {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Sprintf("unknown action %s", action)))
			return
		}

		server, err := controllerClient.FetchServer(context, serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to fetch server %s: %w", serverUUIDAsString, err),
			))

			return
		}

		if server.OperatorRef.Identifier != operatorIdentifier {
			_ = context.Error(response.RestErrorFromDescription(
				http.StatusBadRequest,
				fmt.Sprintf("server %s is not managed by operator %s", serverUUID.String(), operatorIdentifier),
			))

			return
		}

		switch action {
		case networkmodel.Start:
		case networkmodel.Stop:
		case networkmodel.Restart:
		case networkmodel.UpgradeDeployment:
		default:
			_ = context.Error(response.RestErrorFromDescription(http.StatusInternalServerError, fmt.Sprintf("unhandled action %s", action)))
		}
	}
}

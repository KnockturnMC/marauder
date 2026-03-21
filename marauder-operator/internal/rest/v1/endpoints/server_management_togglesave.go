package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/controller"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
	"github.com/knockturnmc/marauder/marauder-operator/pkg/manager"
	"github.com/knockturnmc/marauder/marauder-proto/src/main/golang/marauderpb"
)

func ServerManagementToggleSave(
	operatorIdentifier string,
	controllerClient controller.Client,
	serverManager manager.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUIDAsString := context.Param("uuid")
		serverUUID, err := uuid.Parse(serverUUIDAsString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		var body networkmodel.ManagementToggleSaveBody
		if err := context.BindJSON(&body); err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("could not bindy body: %w", err)))
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

		var manageResponse marauderpb.ServerToggleSaveRequest_Response
		if err := serverManager.ExchangeManagementMessage(
			context,
			server,
			marauderpb.ServerToggleSaveRequest_builder{Save: &body.ShouldSave}.Build(),
			&manageResponse,
		); err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to post toggle to server %s: %w", serverUUIDAsString, err),
			))

			return
		}

		context.Status(http.StatusOK)
	}
}

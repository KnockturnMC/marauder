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

func ServerManagementToggleSave(
	serverManager manager.Manager,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		server, err := getServerModelKey(context)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, err))
			return
		}

		var body networkmodel.ManagementToggleSaveBody
		if err := context.BindJSON(&body); err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("could not bindy body: %w", err)))
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
				fmt.Errorf("failed to post toggle to server %s: %w", server.UUID, err),
			))

			return
		}

		context.Status(http.StatusOK)
	}
}

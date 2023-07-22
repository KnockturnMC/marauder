package endpoints

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeploymentServerPatch creates the patch that may be used to update the is state of servers.
func DeploymentServerPatch(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUID := context.Param("server")
		serverID, err := uuid.Parse(serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		state := networkmodel.ServerStateType(strings.ToUpper(context.Param("state")))
		if !networkmodel.KnownServerStateType(state) || (state != networkmodel.IS && state != networkmodel.TARGET) {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Sprintf(
				"uploading for state '%s' is not supported", state,
			)))

			return
		}

		updateRequest := networkmodel.UpdateServerStateRequest{}
		if err := context.Bind(&updateRequest); err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Errorf("failed to bind body: %w", err).Error()))
			return
		}

		if err := updateRequest.CheckFilled(); err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, err.Error()))
			return
		}

		deployment, err := access.UpdateDeployment(context, db, networkmodel.ServerArtefactStateModel{
			Server:             serverID,
			ArtefactIdentifier: updateRequest.ArtefactIdentifier,
			ArtefactUUID:       updateRequest.ArtefactUUID,
			DefinitionDate:     time.Now(),
			Type:               state,
		})
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(
				http.StatusInternalServerError,
				fmt.Errorf("failed to update deployment for server %s/%s: %w", serverID, updateRequest.ArtefactIdentifier, err)),
			)

			return
		}

		context.JSONP(http.StatusOK, deployment)
	}
}

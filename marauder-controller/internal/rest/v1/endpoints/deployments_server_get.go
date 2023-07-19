package endpoints

import (
	"fmt"
	"net/http"
	"strings"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/networkmodel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeploymentServerGet creates the get endpoint that fetches the deployments of a server.
func DeploymentServerGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUID := context.Param("server")
		state := networkmodel.ServerStateType(strings.ToUpper(context.Param("state")))
		if state == "" {
			state = networkmodel.IS
		}

		if !networkmodel.KnownServerStateType(state) {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, fmt.Sprintf("unknown state %s", state)))
			return
		}

		serverID, err := uuid.Parse(serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		updates, err := access.FetchServerArtefactsByState(context, db, serverID, state)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to fetch updates: %w", err)))
			return
		}

		context.JSONP(http.StatusOK, updates)
	}
}

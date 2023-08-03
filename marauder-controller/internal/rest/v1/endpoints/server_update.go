package endpoints

import (
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerUpdate creates the get endpoint that computes outstanding updates for a server.
func ServerUpdate(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUID := context.Param("uuid")
		serverID, err := uuid.Parse(serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		updates, err := access.FindServerTargetStateMissMatch(context, db, serverID)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to fetch updates: %w", err)))
			return
		}

		context.JSONP(http.StatusOK, updates)
	}
}

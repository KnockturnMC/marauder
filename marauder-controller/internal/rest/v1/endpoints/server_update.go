package endpoints

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
)

// ServerUpdate creates the get endpoint that computes outstanding updates for a server.
func ServerUpdate(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUID := context.Param("uuid")
		requiresRestartString, found := context.GetQuery("requiresRestart")
		if !found {
			requiresRestartString = "true"
		}

		serverID, err := uuid.Parse(serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		requiresRestart, err := strconv.ParseBool(requiresRestartString)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse requiresRestart in query params"))
			return
		}

		updates, err := access.FindServerTargetStateMissMatch(context, db, serverID, requiresRestart)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to fetch updates: %w", err)))
			return
		}

		context.JSONP(http.StatusOK, updates)
	}
}

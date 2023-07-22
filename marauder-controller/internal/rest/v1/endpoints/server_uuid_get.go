package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServerUUIDGet creates the get endpoint that may be used to fetch a specific server based on its uuid.
func ServerUUIDGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		serverUUID := context.Param("uuid")
		serverID, err := uuid.Parse(serverUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		server, err := access.FetchServer(context, db, serverID)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("failed to find server %s", serverID.String())},
			}, fmt.Errorf("failed to fetch server: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, server)
	}
}

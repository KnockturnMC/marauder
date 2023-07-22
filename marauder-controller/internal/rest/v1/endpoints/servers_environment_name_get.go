package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
)

// ServersEnvironmentNameGet creates the get endpoint that may be used to fetch servers based on their environment and name.
func ServersEnvironmentNameGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		environment := context.Param("environment")
		name := context.Param("name")

		server, err := access.FetchServerByNameAndEnv(context, db, name, environment)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("failed to query servers for env %s", environment)},
			}, fmt.Errorf("failed to fetch servers: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, server)
	}
}

package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
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
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("failed to query server %s/%s", environment, name)},
			}, fmt.Errorf("failed to fetch servers: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, server)
	}
}

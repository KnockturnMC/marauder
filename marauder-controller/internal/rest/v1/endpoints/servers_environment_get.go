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

// ServersEnvironmentGet creates the get endpoint that may be used to fetch servers based on their environment.
func ServersEnvironmentGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		environment := context.Param("environment")

		servers, err := access.FetchServersByEnvironment(context, db, environment)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: "failed to query servers for env %s" + environment},
			}, fmt.Errorf("failed to fetch servers: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, servers)
	}
}

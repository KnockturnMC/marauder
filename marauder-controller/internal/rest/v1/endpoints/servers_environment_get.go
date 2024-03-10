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

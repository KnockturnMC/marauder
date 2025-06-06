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

// ArtefactIdentifierVersionGet creates the get endpoint to query an artefact based on its identifier and version.
func ArtefactIdentifierVersionGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := context.Param("identifier")
		version := context.Param("version")

		artefact, err := access.FetchArtefact(context, db, identifier, version)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("failed to find artefact %s:%s", identifier, version)},
			}, fmt.Errorf("failed to fetch artefact: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, artefact)
	}
}

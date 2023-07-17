package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
)

// ArtefactByIdentifierAndVersion creates the get endpoint to query an artefact based on its identifier and version.
func ArtefactByIdentifierAndVersion(
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

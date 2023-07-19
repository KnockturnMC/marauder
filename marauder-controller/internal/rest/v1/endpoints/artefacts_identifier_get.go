package endpoints

import (
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
)

// ArtefactsIdentifierGet creates the get endpoint to query all artefacts for a specific identifier.
func ArtefactsIdentifierGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		identifier := context.Param("identifier")

		artefacts, err := access.FetchArtefactVersions(context, db, identifier)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to fetch artefacts from db: %w", err)))
			return
		}

		context.JSONP(http.StatusOK, artefacts)
	}
}

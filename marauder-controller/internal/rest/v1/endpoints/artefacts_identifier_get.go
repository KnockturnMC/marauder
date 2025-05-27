package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
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

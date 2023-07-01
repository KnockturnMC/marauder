package endpoints

import (
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
)

// ArtefactUpload creates the upload endpoint to which new artefact can be uploaded.
func ArtefactUpload(db *sqlm.DB) gin.HandlerFunc {
	return func(context *gin.Context) {
		_, err := context.FormFile("artefact")
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("failed to read uploaded file artefact: %w", err)))
			return
		}
	}
}

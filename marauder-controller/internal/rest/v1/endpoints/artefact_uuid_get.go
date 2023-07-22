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

// ArtefactUUIDGet creates the get endpoint that may be used to fetch a specific artefact based on its uuid.
func ArtefactUUIDGet(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		artefactUUID := context.Param("uuid")
		artefactID, err := uuid.Parse(artefactUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		artefact, err := access.FetchArtefactByUUID(context, db, artefactID)
		if err != nil {
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("failed to find artefact %s", artefactID.String())},
			}, fmt.Errorf("failed to fetch artefact: %w", err)))

			return
		}

		context.JSONP(http.StatusOK, artefact)
	}
}

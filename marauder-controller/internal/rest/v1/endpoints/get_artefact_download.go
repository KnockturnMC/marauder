package endpoints

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ArtefactByUUIDDownload creates the get endpoint that may be used to download an artefact from the controller.
func ArtefactByUUIDDownload(
	db *sqlm.DB,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		artefactUUID := context.Param("uuid")
		artefactID, err := uuid.Parse(artefactUUID)
		if err != nil {
			_ = context.Error(response.RestErrorFromDescription(http.StatusBadRequest, "could not parse uuid in url params"))
			return
		}

		tarball, err := access.FetchArtefactTarball(context, db, artefactID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = context.Error(response.RestErrorFromDescription(
					http.StatusNotFound,
					fmt.Sprintf("artefact with uuid %s not found", artefactID.String()),
				))
			} else {
				_ = context.Error(response.RestErrorFromErr(
					http.StatusInternalServerError,
					fmt.Errorf("failed to fetch artefact %s: %w", artefactID.String(), err),
				))
			}
			return
		}

		context.Header("Content-Disposition", "attachment; filename="+strconv.Quote(fmt.Sprintf("%s-%s.tar.gz", tarball.Identifier, tarball.Version)))
		context.Data(http.StatusOK, "application/octet-stream", tarball.TarballBlob)
	}
}

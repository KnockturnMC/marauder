package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ArtefactUUIDDownloadGet creates the get endpoint that may be used to download an artefact from the controller.
func ArtefactUUIDDownloadGet(
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
			_ = context.Error(response.RestErrorFromKnownErr(map[error]response.KnownErr{
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("could not find artefact %s", artefactID.String())},
			}, fmt.Errorf("failed to fetch artefact: %w", err)))

			return
		}

		context.Header("Content-Disposition", "attachment; filename="+strconv.Quote(fmt.Sprintf("%s-%s.tar.gz", tarball.Identifier, tarball.Version)))
		context.Data(http.StatusOK, "application/octet-stream", tarball.TarballBlob)
	}
}

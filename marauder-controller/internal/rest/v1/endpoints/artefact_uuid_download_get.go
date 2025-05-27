package endpoints

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/rest/response"
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
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: "could not find artefact " + artefactID.String()},
			}, fmt.Errorf("failed to fetch artefact: %w", err)))

			return
		}

		context.Header("Content-Disposition", "attachment; filename="+strconv.Quote(fmt.Sprintf("%s-%s.tar.gz", tarball.Identifier, tarball.Version)))
		context.Data(http.StatusOK, "application/octet-stream", tarball.TarballBlob)
	}
}

package endpoints

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	artefactlib "gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/rest/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ArtefactUUIDDownloadManifestGet creates the get endpoint that may be used to download the manifest of an artefact from the controller.
func ArtefactUUIDDownloadManifestGet(
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

		manifest, err := readManifestFromTarball(&tarball)
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to read amnifest from artefact: %w", err)))
			return
		}

		// Tarball simply did not contain manifest
		if manifest == nil {
			_ = context.Error(response.RestErrorFromDescription(
				http.StatusInternalServerError,
				fmt.Sprintf("artefact %s did not contain manifest file!", artefactID),
			))

			return
		}

		context.JSON(http.StatusOK, manifest)
	}
}

// readManifestFromTarball reads the manifest from the passed artefact model with binary.
func readManifestFromTarball(binary *networkmodel.ArtefactModelWithBinary) (*artefactlib.Manifest, error) {
	byteReader := bytes.NewBuffer(binary.TarballBlob)

	manifest, err := artefact.ReadManifestFromTarball(byteReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return manifest, nil
}

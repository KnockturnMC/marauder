package endpoints

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg"
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
				sql.ErrNoRows: {ResponseCode: http.StatusNotFound, Description: fmt.Sprintf("could not find artefact %s", artefactID.String())},
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
	gzipReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() { _ = gzipReader.Close() }()

	tarballReader := tar.NewReader(gzipReader)
	var manifestPtr *artefactlib.Manifest
	for {
		nextHeader, err := tarballReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, fmt.Errorf("failed to read next tarball header: %w", err)
			}
		}

		if nextHeader.Name != pkg.ManifestFileName {
			continue
		}

		manifestAsBytes, err := io.ReadAll(tarballReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest bytes from tarball: %w", err)
		}

		manifest := artefactlib.Manifest{}
		if err := json.Unmarshal(manifestAsBytes, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest from json: %w", err)
		}

		manifestPtr = &manifest

		break
	}

	return manifestPtr, nil
}

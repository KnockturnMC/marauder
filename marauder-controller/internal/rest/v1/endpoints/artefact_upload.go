package endpoints

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"gitea.knockturnmc.com/marauder/controller/pkg/artefact"

	"gitea.knockturnmc.com/marauder/controller/internal/rest/v1/response"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/gin-gonic/gin"
)

// ArtefactUpload creates the upload endpoint to which new artefact can be uploaded.
func ArtefactUpload(
	db *sqlm.DB,
	validator artefact.Validator,
) gin.HandlerFunc {
	return func(context *gin.Context) {
		pathToArtefact, err := saveUploadInto(context, "artefact", os.TempDir()+"/marauder", "artefact-*.tar.gz")
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to save artefact file: %w", err)))
			return
		}

		defer func() { _ = os.Remove(pathToArtefact) }()

		pathToSignature, err := saveUploadInto(context, "signature", os.TempDir()+"/marauder", "signature-*.tar.gz.sig")
		if err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusInternalServerError, fmt.Errorf("failed to save signature file: %w", err)))
			return
		}

		defer func() { _ = os.Remove(pathToSignature) }()

		validationResult := <-validator.SubmitArtefact(pathToArtefact, pathToSignature)
		if validationResult.Err != nil {
			_ = context.Error(response.RestErrorFromErr(http.StatusBadRequest, fmt.Errorf("uploaded artefact did not validate: %w", validationResult.Err)))
			return
		}

		context.String(http.StatusOK, "ok!")
	}
}

// saveUploadInto saves the artefact passed into the parent path using the passed pattern as a file name
// which will be expanded by os.CreateTemp.
// The method returns the full path to the saved file.
func saveUploadInto(context *gin.Context, formName string, parentPath string, pattern string) (string, error) {
	header, err := context.FormFile(formName)
	if err != nil {
		return "", fmt.Errorf("failed to find form file for form name %s: %w", formName, err)
	}

	open, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open multipart upload: %w", err)
	}

	defer func() { _ = open.Close() }()

	if err := os.Mkdir(parentPath, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("failed to create parent directory for temporary file: %w", err)
		}
	}

	temp, err := os.CreateTemp(parentPath, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create and open temporary file on disk: %w", err)
	}

	defer func() { _ = temp.Close() }()

	if _, err := io.Copy(temp, open); err != nil {
		return "", fmt.Errorf("failed to write uploaded file stream to temporary file %s: %w", temp.Name(), err)
	}

	return temp.Name(), nil
}

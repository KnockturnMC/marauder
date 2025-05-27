package manager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/knockturnmc/marauder/marauder-lib/pkg"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/fileeq"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/filemodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// ErrFileUnequal is yielded back if a file on disk does not match the file in the artefact that should be deployed.
var ErrFileUnequal = errors.New("files do not match")

// validateOldDeploymentFilesOnDisk validates an old deployment on the disk.
func (d DockerBasedManager) validateOldDeploymentFilesOnDisk(
	oldArtefact filemodel.Manifest,
	oldArtefactOnDisk string,
	serverFolderLocation string,
	fileEqualityRegistry fileeq.FileEqualityRegistry,
) error {
	tarballReader, err := utils.NewFriendlyTarballReaderFromPath(oldArtefactOnDisk)
	if err != nil {
		return fmt.Errorf("failed to open old artefact tarball: %w", err)
	}

	defer func() { _ = tarballReader.Close(true) }()

	fileReferenceMap := oldArtefact.Files.MatchedFilesToReferenceMap()
	for {
		header, err := tarballReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("failed to read next header from artefact: %w", err)
		}

		fileReference, found := fileReferenceMap[header.Name]
		if !found {
			continue
		}

		deployment := utils.OrElse(fileReference.Deployment, filemodel.FileDeployment{})
		fileEqualityIdentifier := utils.OrElse(deployment.EqualityProvider, "hash")
		fileEquality, found := fileEqualityRegistry[fileEqualityIdentifier]
		if !found {
			return fmt.Errorf("%s is an unknown file equality: %w", fileEqualityIdentifier, fileeq.ErrUnknownFileEquality)
		}

		filePathInServerFolder, _ := strings.CutPrefix(header.Name, pkg.FileParentDirectoryInArtefact)
		fullyQualifiedFilePath := path.Join(serverFolderLocation, path.Clean(filePathInServerFolder))

		fileOnDisk, err := os.Open(filepath.Clean(fullyQualifiedFilePath))
		if err != nil {
			return fmt.Errorf("failed to open expected file %s: %w", fullyQualifiedFilePath, err)
		}

		equals, err := fileEquality.Equals(fileOnDisk, tarballReader)
		if err != nil {
			_ = fileOnDisk.Close()
			return fmt.Errorf("failed to compare file %s with expected state: %w", filePathInServerFolder, err)
		}

		_ = fileOnDisk.Close()

		if !equals {
			return fmt.Errorf("file %s [%s] did not match expected state: %w", filePathInServerFolder, fileEqualityIdentifier, ErrFileUnequal)
		}
	}
}

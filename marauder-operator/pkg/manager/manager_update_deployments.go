package manager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Goldziher/go-utils/sliceutils"

	"github.com/google/uuid"

	"github.com/Goldziher/go-utils/maputils"
	"golang.org/x/exp/slices"

	"gitea.knockturnmc.com/marauder/controller/pkg/artefact"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/lib/pkg"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
	"github.com/docker/docker/client"
)

// ErrServerRunning is returned by UpdateDeployments if the server is running.
var ErrServerRunning = errors.New("server is running")

func (d DockerBasedManager) UpdateDeployments(ctx context.Context, serverModel networkmodel.ServerModel) error {
	_, err := d.retrieveContainerInfo(ctx, serverModel)
	if err == nil {
		return fmt.Errorf("server %s is running: %w", serverModel.UUID.String(), ErrServerRunning)
	} else if !utils.CheckDockerError(err, client.IsErrNotFound) {
		return fmt.Errorf("failed to fetch container info for %s: %w", serverModel.UUID.String(), err)
	}

	missmatches, err := d.ControllerClient.FetchMissmatchesFor(ctx, serverModel.UUID)
	if err != nil {
		return fmt.Errorf("failed to fetch missmatches for %s: %w", serverModel.UUID, err)
	}

	for _, update := range missmatches {
		if err := d.updateSingleDeployment(ctx, serverModel, update); err != nil {
			return fmt.Errorf("failed to update %s on %s: %w", update.ArtefactIdentifier, serverModel.UUID.String(), err)
		}
	}

	return nil
}

// updateSingleDeployment updates a single deployment on the server.
//
//nolint:cyclop
func (d DockerBasedManager) updateSingleDeployment(
	ctx context.Context,
	serverModel networkmodel.ServerModel,
	update networkmodel.ArtefactVersionMissmatch,
) error {
	serverFolderLocation, err := d.computeServerFolderLocation(serverModel)
	if err != nil {
		return fmt.Errorf("failed to compute server folder location: %w", err)
	}

	artefactToInstall := update.Missmatch.ArtefactToInstall()
	artefactToUninstall := update.Missmatch.ArtefactToUninstall()
	var (
		artefactToUninstallManifest filemodel.Manifest
		artefactToInstallOnDisk     string
	)

	if artefactToUninstall != nil {
		artefactToUninstallManifest, err = d.ControllerClient.FetchManifest(ctx, artefactToUninstall.Artefact)
		if err != nil {
			return fmt.Errorf("failed to fetch old manifest: %w", err)
		}

		if err := d.validateOldDeploymentHashOnDisk(artefactToUninstallManifest, serverFolderLocation); err != nil {
			return fmt.Errorf("failed to validate old deployment hashes: %w", err)
		}
	}

	if artefactToInstall != nil {
		artefactToInstallOnDisk, err = d.ControllerClient.DownloadArtefact(ctx, artefactToInstall.Artefact)
		if err != nil {
			return fmt.Errorf("failed to download target artefact: %w", err)
		}
	}

	if artefactToUninstall != nil {
		// Delete old artefact files after downloading the new one to fail before moving the server into a non-start-able state
		// This needs further improvements down the line to do a proper rollback, for now this should be fine.
		if err := d.deleteOldArtefact(artefactToUninstallManifest, serverFolderLocation, true); err != nil {
			return fmt.Errorf("failed to delete old artefact from server folder: %w", err)
		}
	}

	var artefactToInstallUUID *uuid.UUID
	if artefactToInstall != nil {
		// Extract the new artefact to the server directory
		if err := d.unpackArtefactIntoServer(artefactToInstallOnDisk, serverFolderLocation); err != nil {
			var oldArtefactUUID *uuid.UUID
			if artefactToUninstall != nil {
				oldArtefactUUID = &artefactToUninstall.Artefact
			}

			// Unpacking failed, undo all potentially unpacked files.
			if err := d.rollbackFailedArtefactInstall(ctx, artefactToInstallOnDisk, oldArtefactUUID, serverFolderLocation); err != nil {
				return fmt.Errorf("failed to rollback failed artefact installation: %w", err)
			}

			return fmt.Errorf("failed to unpack new artefact: %w", err)
		}

		artefactToInstallUUID = &artefactToInstall.Artefact
	}

	if err := d.ControllerClient.UpdateState(ctx, serverModel.UUID, networkmodel.IS, networkmodel.UpdateServerStateRequest{
		ArtefactIdentifier: update.ArtefactIdentifier,
		ArtefactUUID:       artefactToInstallUUID,
	}); err != nil {
		return fmt.Errorf("failed to update controllers is state for server: %w", err)
	}

	return nil
}

// validateOldDeploymentHashOnDisk validates an old deployment on the disk.
func (d DockerBasedManager) validateOldDeploymentHashOnDisk(oldArtefact filemodel.Manifest, serverFolderLocation string) error {
	// Check hashes of old deployment, ensuring that the state of the local server dir matches the expected one.
	// This could differ if some system developer edited files in the server environment.
	for filePathWithPrefix, fileHashAsString := range oldArtefact.Hashes {
		filePathWithoutPrefix, _ := strings.CutPrefix(filePathWithPrefix, pkg.FileParentDirectoryInArtefact)

		fileToCheck, err := os.Open(path.Join(serverFolderLocation, filePathWithoutPrefix))
		if err != nil {
			return fmt.Errorf("failed to open file for hash validation: %w", err)
		}

		sha256, err := utils.ComputeSha256(fileToCheck)
		if err != nil {
			_ = fileToCheck.Close() // Close file, not done via defer as we are in a for loop
			return fmt.Errorf("failed to compute sha256 hash for %s: %w", fileToCheck.Name(), err)
		}

		_ = fileToCheck.Close() // close file for good after sha was computed.
		foundHashAsHex := hex.EncodeToString(sha256)
		if fileHashAsString != foundHashAsHex {
			return fmt.Errorf("file %s has hash %s, expected %s: %w", fileToCheck.Name(), foundHashAsHex, fileHashAsString, artefact.ErrHashMismatch)
		}
	}

	return nil
}

// deleteOldArtefact deletes an old artefact in the server folder.
func (d DockerBasedManager) deleteOldArtefact(
	oldArtefact filemodel.Manifest,
	serverFolderLocation string,
	errorOnMissingFiles bool,
) error {
	relativePotentiallyEmptyParentDirsAsMap, err := d.deleteOldFilesAndYieldParents(oldArtefact, serverFolderLocation, errorOnMissingFiles)
	if err != nil {
		return fmt.Errorf("failed to delete old files: %w", err)
	}

	// Sort to depth
	potentiallyEmptyDirs := maputils.Keys(relativePotentiallyEmptyParentDirsAsMap)
	slices.SortFunc(potentiallyEmptyDirs, func(a, b string) int {
		return strings.Count(b, "/") - strings.Count(a, "/")
	})

	// Iterate and remove is possible.
	for _, dir := range potentiallyEmptyDirs {
		dirOutput, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("failed to read potentially empty parent dir %s: %w", dir, err)
		}

		if len(dirOutput) == 0 {
			if err := os.Remove(path.Join(serverFolderLocation, dir)); err != nil {
				return fmt.Errorf("failed to remove empty parent dir %s: %w", dir, err)
			}
		}
	}

	return nil
}

// deleteOldFilesAndYieldParents deletes all old files of a manifest and yields back a map of all parent dirs that might need to be deleted if empty.
func (d DockerBasedManager) deleteOldFilesAndYieldParents(
	oldArtefact filemodel.Manifest,
	serverFolderLocation string,
	errorOnMissingFiles bool,
) (map[string]bool, error) {
	relativePotentiallyEmptyParentDirsAsMap := make(map[string]bool)

	// Delete all files
	for filePathWithPrefix := range oldArtefact.Hashes {
		filePathWithoutPrefix, _ := strings.CutPrefix(filePathWithPrefix, pkg.FileParentDirectoryInArtefact)
		cleanedFilePathWithoutPrefix := path.Clean(filePathWithoutPrefix)

		fullFilePath := path.Join(serverFolderLocation, cleanedFilePathWithoutPrefix)
		if err := os.Remove(fullFilePath); err != nil && os.IsNotExist(err) && errorOnMissingFiles {
			return relativePotentiallyEmptyParentDirsAsMap, fmt.Errorf("failed to delete file %s: %w", filePathWithPrefix, err)
		}

		for {
			cleanedFilePathWithoutPrefix = path.Dir(cleanedFilePathWithoutPrefix)
			if cleanedFilePathWithoutPrefix == "." {
				break
			}

			relativePotentiallyEmptyParentDirsAsMap[cleanedFilePathWithoutPrefix] = true
		}
	}

	// Collect protected dirs from file references
	protectedDirsSlice := sliceutils.Map(oldArtefact.Files, func(value filemodel.FileReference, index int, slice []filemodel.FileReference) string {
		if strings.HasSuffix(value.Target, "/") {
			return strings.TrimSuffix(value.Target, "/") // Return the entire directory (without trailing slash)
		}

		return filepath.Dir(value.Target) // It's a file, return its dir name. Does not have a trailing slash either
	})
	protectedDirsSlice = sliceutils.Unique(protectedDirsSlice)

	// Filter out protected dirs.
	relativePotentiallyEmptyParentDirsAsMap = maputils.Filter(relativePotentiallyEmptyParentDirsAsMap, func(key string, value bool) bool {
		for _, protectedDir := range protectedDirsSlice {
			if strings.Contains(protectedDir, key) {
				return false
			}
		}

		return true
	})

	return relativePotentiallyEmptyParentDirsAsMap, nil
}

// unpackArtefactIntoServer unpacks the passed artefact into the server.
func (d DockerBasedManager) unpackArtefactIntoServer(artefactPath string, serverFolderLocation string) error {
	plainArtefactReader, err := os.Open(artefactPath)
	if err != nil {
		return fmt.Errorf("failed to open artefact tar: %w", err)
	}

	defer func() { _ = plainArtefactReader.Close() }()

	gzippedReader, err := gzip.NewReader(plainArtefactReader)
	if err != nil {
		return fmt.Errorf("failed to create gzipped reader for artefact: %w", err)
	}

	defer func() { _ = gzippedReader.Close() }()

	tarballReader := tar.NewReader(gzippedReader)
	for {
		tarballHeader, err := tarballReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to read next tarball header: %w", err)
		}

		// Something not in the files/ top level dir.
		if !strings.HasPrefix(tarballHeader.Name, pkg.FileParentDirectoryInArtefact) {
			continue
		}

		if err := d.extractFileToServer(tarballHeader, tarballReader, serverFolderLocation); err != nil {
			return fmt.Errorf("failed to extract tar file: %w", err)
		}
	}

	return nil
}

// extractFileToServer extracts the file in the tarball to the server folder location.
func (d DockerBasedManager) extractFileToServer(tarballHeader *tar.Header, tarballReader *tar.Reader, serverFolderLocation string) error {
	filePathInServerFolder, _ := strings.CutPrefix(tarballHeader.Name, pkg.FileParentDirectoryInArtefact)
	cleanedFilePathInServerFolder := path.Clean(filePathInServerFolder)
	filePathOnSystem := path.Join(serverFolderLocation, cleanedFilePathInServerFolder)

	if err := os.MkdirAll(path.Dir(filePathOnSystem), 0o700); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", filePathOnSystem, err)
	}

	targetFileOnSystem, err := os.Create(filePathOnSystem)
	if err != nil {
		return fmt.Errorf("failed to open output file %s: %w", filePathOnSystem, err)
	}

	if _, err := io.CopyN(targetFileOnSystem, tarballReader, tarballHeader.Size); err != nil {
		_ = targetFileOnSystem.Close()
		return fmt.Errorf("failed to copy over tarball content to disk file %s: %w", filePathOnSystem, err)
	}

	_ = targetFileOnSystem.Close()

	if d.FolderOwner != nil {
		chownedFilePath := cleanedFilePathInServerFolder
		for chownedFilePath != "." {
			if err := os.Chown(path.Join(serverFolderLocation, chownedFilePath), d.FolderOwner.UID, d.FolderOwner.GID); err != nil {
				return fmt.Errorf("failed to chown deployed file %s: %w", chownedFilePath, err)
			}

			chownedFilePath = path.Dir(chownedFilePath)
		}
	}
	return nil
}

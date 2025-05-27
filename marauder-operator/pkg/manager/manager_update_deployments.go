package manager

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Goldziher/go-utils/maputils"
	"github.com/Goldziher/go-utils/sliceutils"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/filemodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/sirupsen/logrus"
)

// ErrServerRunning is returned by UpdateDeployments if the server is running.
var ErrServerRunning = errors.New("server is running")

func (d DockerBasedManager) UpdateDeployments(
	ctx context.Context,
	serverModel networkmodel.ServerModel,
	requiresRestart bool,
	failOnUnexpectedOldFilesOnDisk bool,
) error {
	if requiresRestart {
		_, err := d.retrieveContainerInfo(ctx, serverModel)
		if err == nil {
			return fmt.Errorf("server %s is running: %w", serverModel.UUID.String(), ErrServerRunning)
		} else if !utils.CheckDockerError(err, client.IsErrNotFound) {
			return fmt.Errorf("failed to fetch container info for %s: %w", serverModel.UUID.String(), err)
		}
	}

	missmatches, err := d.ControllerClient.FetchMissmatchesFor(ctx, serverModel.UUID, requiresRestart)
	if err != nil {
		return fmt.Errorf("failed to fetch missmatches for %s: %w", serverModel.UUID, err)
	}

	for _, update := range missmatches {
		if err := d.updateSingleDeployment(ctx, serverModel, update, failOnUnexpectedOldFilesOnDisk); err != nil {
			return fmt.Errorf("failed to update %s on %s: %w", update.ArtefactIdentifier, serverModel.UUID.String(), err)
		}

		logrus.Info("upgraded deployment ", update.ArtefactIdentifier, " on server ", serverModel.Environment, "/", serverModel.Name)
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
	failOnUnexpectedOldFilesOnDisk bool,
) error {
	serverFolderLocation, err := d.computeServerFolderLocation(serverModel)
	if err != nil {
		return fmt.Errorf("failed to compute server folder location: %w", err)
	}

	artefactToInstall := update.Missmatch.ArtefactToInstall()
	artefactToUninstall := update.Missmatch.ArtefactToUninstall()
	var (
		artefactToUninstallManifest filemodel.Manifest
		artefactToUninstallOnDisk   string
		artefactToInstallOnDisk     string
	)

	if artefactToUninstall != nil {
		artefactToUninstallOnDisk, err = d.ControllerClient.DownloadArtefact(ctx, artefactToUninstall.Artefact)
		if err != nil {
			return fmt.Errorf("failed to fetch old artefact to disk: %w", err)
		}

		artefactToUninstallManifest, err = d.ControllerClient.FetchManifest(ctx, artefactToUninstall.Artefact)
		if err != nil {
			return fmt.Errorf("failed to fetch old artefact manifest: %w", err)
		}

		if err := d.validateOldDeploymentFilesOnDisk(
			artefactToUninstallManifest,
			artefactToUninstallOnDisk,
			serverFolderLocation,
			d.FileEqualityRegistry,
		); err != nil {
			if failOnUnexpectedOldFilesOnDisk {
				return fmt.Errorf("failed to validate old deployment hashes: %w", err)
			} else {
				logrus.Warning(
					"found unexpected file of ", update.ArtefactIdentifier, " on server ", serverModel.Environment, "/", serverModel.Name+
						" while upgrading from version ", artefactToUninstall.Version, " ", err,
				)
			}
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
		dirOutput, err := os.ReadDir(path.Join(serverFolderLocation, dir))
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
	for filePathWithPrefix := range oldArtefact.Files.MatchedFilesToReferenceMap() {
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
	tarballReader, err := utils.NewFriendlyTarballReaderFromPath(artefactPath)
	if err != nil {
		return fmt.Errorf("failed to open artefact tar: %w", err)
	}

	defer func() { _ = tarballReader.Close(true) }()

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

		if err := d.extractFileToServer(tarballHeader, tarballReader.Reader, serverFolderLocation); err != nil {
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

	targetFileOnSystem, err := os.Create(filepath.Clean(filePathOnSystem))
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

package manager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/Goldziher/go-utils/maputils"
	"golang.org/x/exp/slices"
	"io"
	"os"
	"path"
	"strings"

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

	updates, err := d.ControllerClient.FetchUpdatesFor(ctx, serverModel.UUID)
	if err != nil {
		return fmt.Errorf("failed to fetch updates for %s: %w", serverModel.UUID, err)
	}

	for _, update := range updates {
		if err := d.updateSingleDeployment(ctx, serverModel, update); err != nil {
			return fmt.Errorf("failed to update %s on %s: %w", update.ArtefactIdentifier, serverModel.UUID.String(), err)
		}
	}

	return nil
}

// updateSingleDeployment updates a single deployment on the server.
func (d DockerBasedManager) updateSingleDeployment(
	ctx context.Context,
	serverModel networkmodel.ServerModel,
	update networkmodel.VersionDiff,
) error {
	oldArtefact, err := d.ControllerClient.FetchManifest(ctx, update.IsArtefact)
	if err != nil {
		return fmt.Errorf("failed to fetch old manifest: %w", err)
	}

	serverFolderLocation, err := d.computeServerFolderLocation(serverModel)
	if err != nil {
		return fmt.Errorf("failed to compute server folder location: %w", err)
	}

	if err := d.validateOldDeploymentHashOnDisk(oldArtefact, serverFolderLocation); err != nil {
		return fmt.Errorf("failed to validate old deployment hashes: %w", err)
	}

	targetArtefact, err := d.ControllerClient.DownloadArtefact(ctx, update.TargetArtefact)
	if err != nil {
		return fmt.Errorf("failed to download target artefact: %w", err)
	}

	// Delete old artefact files after downloading the new one to fail before moving the server into a non-start-able state
	// This needs further improvements down the line to do a proper rollback, for now this should be fine.
	if err := d.deleteOldArtefact(oldArtefact, serverFolderLocation); err != nil {
		return fmt.Errorf("failed to delete old artefact from server folder: %w", err)
	}

	// Extract the new artefact to the server directory
	if err := d.unpackArtefactIntoServer(targetArtefact, serverFolderLocation); err != nil {
		return fmt.Errorf("failed to unpack new artefact: %w", err)
	}

	if err := d.ControllerClient.UpdateIsState(ctx, serverModel.UUID, update.ArtefactIdentifier, update.TargetArtefact); err != nil {
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
func (d DockerBasedManager) deleteOldArtefact(oldArtefact filemodel.Manifest, serverFolderLocation string) error {
	potentiallyEmptyParentDirsAsMap := make(map[string]bool)

	// Delete all files
	for filePathWithPrefix := range oldArtefact.Hashes {
		filePathWithoutPrefix, _ := strings.CutPrefix(filePathWithPrefix, pkg.FileParentDirectoryInArtefact)
		cleanedFilePathWithoutPrefix := path.Clean(filePathWithoutPrefix)

		fullFilePath := path.Join(serverFolderLocation, cleanedFilePathWithoutPrefix)
		if err := os.Remove(fullFilePath); err != nil {
			return fmt.Errorf("failed to delete file %s: %w", filePathWithPrefix, err)
		}

		for {
			cleanedFilePathWithoutPrefix = path.Dir(cleanedFilePathWithoutPrefix)
			if cleanedFilePathWithoutPrefix == "." {
				break
			}

			potentiallyEmptyParentDirsAsMap[path.Join(serverFolderLocation, cleanedFilePathWithoutPrefix)] = true
		}
	}

	// Sort to depth
	potentiallyEmptyDirs := maputils.Keys(potentiallyEmptyParentDirsAsMap)
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
			if err := os.Remove(dir); err != nil {
				return fmt.Errorf("failed to remove empty parent dir %s: %w", dir, err)
			}
		}
	}

	return nil
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

		filePathInServerFolder, _ := strings.CutPrefix(tarballHeader.Name, pkg.FileParentDirectoryInArtefact)
		filePathOnSystem := path.Join(serverFolderLocation, path.Clean(filePathInServerFolder))

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
	}

	return nil
}

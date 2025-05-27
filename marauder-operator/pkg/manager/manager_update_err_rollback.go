package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/artefact"
)

// rollbackFailedArtefactInstall rolls the server back to the previously installed artefact if the installation process failed.
func (d DockerBasedManager) rollbackFailedArtefactInstall(
	ctx context.Context,
	artefactFailedToInstall string,
	oldArtefactUUID *uuid.UUID,
	serverLocation string,
) error {
	failedInstallArtefact, err := os.Open(filepath.Clean(artefactFailedToInstall))
	if err != nil {
		return fmt.Errorf("failed to open failed new artefact: %w", err)
	}

	defer func() { _ = failedInstallArtefact.Close() }()

	manifest, err := artefact.ReadManifestFromTarball(failedInstallArtefact)
	if err != nil || manifest == nil {
		return fmt.Errorf("failed to read manifest from failed new artefact: %w", err)
	}

	// don't error on missing files, the artefact might not have been unpacked completely.
	if err := d.deleteOldArtefact(*manifest, serverLocation, false); err != nil {
		return fmt.Errorf("failed to delete failed new artefact: %w", err)
	}

	if oldArtefactUUID != nil {
		oldArtefactOnDisk, err := d.ControllerClient.DownloadArtefact(ctx, *oldArtefactUUID)
		if err != nil {
			return fmt.Errorf("failed to download old artefact onto disk: %w", err)
		}

		if err := d.unpackArtefactIntoServer(oldArtefactOnDisk, serverLocation); err != nil {
			return fmt.Errorf("failed to unpack old artefact onto server: %w", err)
		}
	}

	return nil
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/client/pkg/builder"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"github.com/spf13/cobra"
)

// RootCommand is the root entry command for the builder tool.
func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "marauder",
		Short: "Marauder is a command line tool capable of constructing artefacts",
		Long: `Marauder is a command line tool capable of packing together a locally defined artefact into a
tarball and uploading said artefact to the marauder controller.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.ReadFile("manifest.json")
			if err != nil {
				return fmt.Errorf("failed to read manifest.json: %w", err)
			}

			var manifest artefact.Manifest
			if err := json.Unmarshal(file, &manifest); err != nil {
				return fmt.Errorf("failed to parse manifest: %w", err)
			}

			if err := builder.CreateArtefactTarball(os.DirFS("."), "artefact.tar.gz", manifest); err != nil {
				return fmt.Errorf("failed to create artefact tarball: %w", err)
			}

			return nil
		},
	}
}

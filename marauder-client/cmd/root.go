package cmd

import (
	"github.com/spf13/cobra"
)

var version = "develop"

// RootCommand is the root entry command for the builder tool.
func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "marauder",
		Short: "Marauder is a command line tool capable of constructing artefacts",
		Long: `Marauder is a command line tool capable of packing together a locally defined artefact into a
tarball and uploading said artefact to the marauder controller.`,
		Version: version,
	}
}

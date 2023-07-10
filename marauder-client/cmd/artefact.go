package cmd

import "github.com/spf13/cobra"

func ArtefactCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "artefact",
		Short: "The parent command for the marauder client interacting with artefacts.",
	}
}

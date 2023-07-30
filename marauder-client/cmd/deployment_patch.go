package cmd

import (
	"github.com/spf13/cobra"
)

func DeploymentPatchCommand(_ *Configuration) *cobra.Command {
	var manifestFileLocation string

	command := &cobra.Command{
		Use:   "patch [flags] <artefactIdentifier> <artefactUUID>",
		Short: "Patches a new deployment target onto ",
		Args:  cobra.ExactArgs(2),
	}
	command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "m", ".marauder.json", "location of the manifest file")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	return command
}

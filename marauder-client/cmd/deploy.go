package cmd

import (
	"github.com/spf13/cobra"
)

// DeployCommand constructs the deploy subcommand.
func DeployCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "deploy",
		Short: "The subcommand to deploy things via marauder",
	}

	return command
}

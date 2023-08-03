package cmd

import (
	"github.com/spf13/cobra"
)

// OperatorCommand constructs the build subcommand.
func BuildCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "build",
		Short: "The subcommand to build things via marauder",
	}

	return command
}

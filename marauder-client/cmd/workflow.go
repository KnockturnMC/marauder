package cmd

import (
	"github.com/spf13/cobra"
)

// WorkflowCommand constructs the workflow subcommand.
func WorkflowCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "workflow",
		Short: "The subcommand to execute workflows via marauder",
	}

	return command
}

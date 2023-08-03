package cmd

import (
	"github.com/spf13/cobra"
)

// PublishCommand constructs the publishing subcommand.
func PublishCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "publish",
		Short: "The subcommand to publish things via marauder",
	}

	return command
}

package cmd

import "github.com/spf13/cobra"

// OperateCommand constructs the operator subcommand.
func OperateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "operate",
		Short: "The subcommand to operate things via marauder",
	}

	return command
}

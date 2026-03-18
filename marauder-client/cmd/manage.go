package cmd

import "github.com/spf13/cobra"

func ManageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "manage",
		Short: "The parent command for the management related actions against live servers.",
	}
}

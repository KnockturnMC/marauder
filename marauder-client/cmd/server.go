package cmd

import "github.com/spf13/cobra"

func ServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "The parent command for the marauder client interacting with servers.",
	}
}

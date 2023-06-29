package cmd

import "github.com/spf13/cobra"

func BuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "The parent command for the marauder client to build something.",
	}
}

package cmd

import "github.com/spf13/cobra"

func GetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "The parent command for the reading from the marauder controller.",
	}
}

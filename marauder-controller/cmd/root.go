package cmd

import "github.com/spf13/cobra"

var version = "develop"

// Version returns the Version of the program.
func Version() string {
	return version
}

// RootCommand generates the root command of marauder controller.
func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "marauderctl",
		Short:   "The marauder controller cli",
		Version: version,
	}
}

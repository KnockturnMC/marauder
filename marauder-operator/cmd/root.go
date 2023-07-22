package cmd

import "github.com/spf13/cobra"

var version = "develop"

// Version returns the Version of the program.
func Version() string {
	return version
}

// RootCommand generates the root command of marauder operator.
func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "marauderop",
		Short:   "The marauder operator cli",
		Version: version,
	}
}

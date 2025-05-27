package main

import (
	"log"

	"github.com/knockturnmc/marauder/marauder-operator/cmd"
)

func main() {
	rootCmd := cmd.RootCommand()
	rootCmd.AddCommand(cmd.ServeCommand())
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

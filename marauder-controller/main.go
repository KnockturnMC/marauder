package main

import (
	"log"

	"gitea.knockturnmc.com/marauder/controller/cmd"
)

func main() {
	rootCmd := cmd.RootCommand()
	rootCmd.AddCommand(cmd.ServeCommand())
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

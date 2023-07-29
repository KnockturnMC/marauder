package main

import (
	"context"
	"log"
	"os"

	"gitea.knockturnmc.com/marauder/client/cmd"
)

func main() {
	root := cmd.RootCommand()
	artefactCommand := cmd.ArtefactCommand()
	artefactCommand.AddCommand(cmd.ArtefactBuildCommand())
	artefactCommand.AddCommand(cmd.ArtefactSignCommand())
	artefactCommand.AddCommand(cmd.ArtefactPublishCommand(context.Background()))
	root.AddCommand(artefactCommand)
	root.SetOut(os.Stdout) // By default, the output should properly be printed to stdout.

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

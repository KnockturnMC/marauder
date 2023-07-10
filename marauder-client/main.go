package main

import (
	"log"

	"gitea.knockturnmc.com/marauder/client/cmd"
)

func main() {
	root := cmd.RootCommand()
	artefactCommand := cmd.ArtefactCommand()
	artefactCommand.AddCommand(cmd.ArtefactBuildCommand())
	artefactCommand.AddCommand(cmd.ArtefactSignCommand())
	artefactCommand.AddCommand(cmd.ArtefactPublishCommand())
	root.AddCommand(artefactCommand)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

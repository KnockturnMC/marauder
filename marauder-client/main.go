package main

import (
	"context"
	"log"
	"os"

	"gitea.knockturnmc.com/marauder/client/cmd"
)

func main() {
	configuration := cmd.DefaultConfiguration()

	context := context.Background()

	root := cmd.RootCommand(&configuration)

	artefactCommand := cmd.ArtefactCommand()
	artefactCommand.AddCommand(cmd.ArtefactBuildCommand())
	artefactCommand.AddCommand(cmd.ArtefactSignCommand(&configuration))
	artefactCommand.AddCommand(cmd.ArtefactPublishCommand(context, &configuration))
	artefactCommand.AddCommand(cmd.ArtefactFetchCommand(context, &configuration))
	root.AddCommand(artefactCommand)

	serverCommand := cmd.ServerCommand()
	serverCommand.AddCommand(cmd.ServersFetchCommand(context, &configuration))
	root.AddCommand(serverCommand)

	root.SetOut(os.Stdout) // By default, the output should properly be printed to stdout.

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

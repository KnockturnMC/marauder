package main

import (
	"log"

	"gitea.knockturnmc.com/marauder/client/cmd"
)

func main() {
	root := cmd.RootCommand()
	build := cmd.BuildCommand()
	build.AddCommand(cmd.BuildArtefactCommand())
	root.AddCommand(build)
	root.AddCommand(cmd.SignCommand())

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

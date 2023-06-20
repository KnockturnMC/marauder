package main

import (
	"log"

	"gitea.knockturnmc.com/marauder/builder/cmd"
)

func main() {
	if err := cmd.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

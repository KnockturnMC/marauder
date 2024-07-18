package main

import (
	"context"
	"log"
	"os"

	"gitea.knockturnmc.com/marauder/client/cmd"
)

func main() {
	configuration := cmd.DefaultConfiguration()

	ctx := context.Background()

	root := cmd.RootCommand(&configuration)

	getCommand := cmd.GetCommand()

	getArtefactCommand := cmd.GetArtefactCommand(ctx, &configuration)
	getArtefactCommand.AddCommand(cmd.GetArtefactManifestCommand(ctx, &configuration))
	getArtefactCommand.AddCommand(cmd.GetArtefactArchiveCommand(ctx, &configuration))
	getCommand.AddCommand(getArtefactCommand)

	getServerCommand := cmd.GetServerCommand(ctx, &configuration)
	getServerCommand.AddCommand(cmd.GetServerStateCommand(ctx, &configuration))
	getCommand.AddCommand(getServerCommand)

	root.AddCommand(getCommand)

	buildCommand := cmd.BuildCommand()
	buildCommand.AddCommand(cmd.BuildArtefactCommand(&configuration))
	root.AddCommand(buildCommand)

	publish := cmd.PublishCommand()
	publish.AddCommand(cmd.PublishArtefactCommand(ctx, &configuration))
	root.AddCommand(publish)

	deployCommand := cmd.DeployCommand()
	deployCommand.AddCommand(cmd.DeployArtefactCommand(ctx, &configuration))
	root.AddCommand(deployCommand)

	operateCommand := cmd.OperateCommand()
	operateCommand.AddCommand(cmd.OperateServerCommand(ctx, &configuration))
	root.AddCommand(operateCommand)

	workflowCommand := cmd.WorkflowCommand()
	workflowCommand.AddCommand(cmd.WorkflowBuildAndDeployCommand(ctx, &configuration))
	root.AddCommand(workflowCommand)

	root.SetOut(os.Stdout) // By default, the output should properly be printed to stdout.

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

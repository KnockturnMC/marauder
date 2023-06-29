package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gitea.knockturnmc.com/marauder/client/pkg/builder"
	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/gonvenience/bunt"
	"github.com/spf13/cobra"
)

type OutputNameData struct {
	Identifier string
	Version    string
}

// BuildArtefactCommand constructs the command logic for the artefact creation.
func BuildArtefactCommand() *cobra.Command {
	var (
		manifestFileLocation string
		tarballName          string
	)

	command := &cobra.Command{
		Use:   "artefact",
		Short: "Builds a marauder artefact into a tarball ready for publishing.",
		Args:  cobra.MaximumNArgs(1),
	}
	command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "m", "manifest.json", "location of the manifest file")
	command.PersistentFlags().StringVarP(&tarballName, "output", "o", "{{.Identifier}}-{{.Version}}-artefact.tar.gz", "name of the output tarball")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		workDirectory := "."
		if len(args) > 0 {
			workDirectory = args[0]
		}

		file, err := os.ReadFile(manifestFileLocation)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", manifestFileLocation, err)
		}

		cmd.Println(bunt.Sprintf("Gray{parsing manifest file %s}", manifestFileLocation))
		var manifest artefact.Manifest
		if err := json.Unmarshal(file, &manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		tarballNameTemplate, err := template.New("").Parse(tarballName)
		if err != nil {
			return fmt.Errorf("failed to parse output name as template: %w", err)
		}

		var stringWriter strings.Builder
		if err := tarballNameTemplate.Execute(&stringWriter, OutputNameData{
			Identifier: manifest.Identifier,
			Version:    manifest.Version,
		}); err != nil {
			return fmt.Errorf("failed to execute template for output file location: %w", err)
		}

		cmd.Println(bunt.Sprintf("Gray{fetching build information from project}"))
		{
			updatedManifest, err := builder.InsertBuildInformation(workDirectory, manifest)
			if err != nil {
				cmd.Println(bunt.Sprintf("Red{failed to parse build information, excluding them: %s}", err.Error()))
			} else {
				manifest = updatedManifest
			}
		}

		cmd.Println(bunt.Sprintf("Gray{creating output artefact tarball *%s*}", stringWriter.String()))
		tarballFileRef, err := os.Create(stringWriter.String())
		if err != nil {
			return fmt.Errorf("failed to open output tarball: %w", err)
		}
		defer utils.SwallowClose(tarballFileRef)

		if err := builder.CreateArtefactTarball(os.DirFS(workDirectory), manifest, tarballFileRef); err != nil {
			return fmt.Errorf("failed to create artefact tarball: %w", err)
		}

		cmd.Println(bunt.Sprintf("LimeGreen{created artefact} *%s* LimeGreen{successfully!}", stringWriter.String()))

		return nil
	}

	return command
}

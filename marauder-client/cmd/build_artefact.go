package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

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

		cmd.Println(bunt.Sprintf("Gray{fetching build information from project}"))
		buildInformation, err := builder.FetchBuildInformation(workDirectory)
		if err != nil {
			cmd.Println(bunt.Sprintf("Red{failed to parse build information, excluding them: %s}", err.Error()))
			timestamp := time.Now()
			buildInformation = artefact.BuildInformation{
				Repository:           "nan",
				Branch:               "nan",
				CommitUser:           "nan",
				CommitEmail:          "nan",
				CommitHash:           "nan",
				CommitMessage:        "nan",
				Timestamp:            timestamp,
				BuildSpecificVersion: "t" + strconv.FormatInt(timestamp.Unix(), 10),
			}
		}

		// Parse the manifest file
		cmd.Println(bunt.Sprintf("Gray{parsing manifest file %s}", manifestFileLocation))

		templatedManifestContent, err := utils.ExecuteStringTemplateToString(string(file), struct {
			Build artefact.BuildInformation
		}{
			Build: buildInformation,
		})
		if err != nil {
			return fmt.Errorf("failed to resolve templates in manifest file: %w", err)
		}

		var manifest artefact.Manifest
		if err := json.Unmarshal([]byte(templatedManifestContent), &manifest); err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
		}

		// Parse the tarball name from the commandline flag
		finalTarballName, err := utils.ExecuteStringTemplateToString(tarballName, OutputNameData{
			Identifier: manifest.Identifier,
			Version:    manifest.Version,
		})
		if err != nil {
			return fmt.Errorf("failed to execute template for tarball output name: %w", err)
		}

		cmd.Println(bunt.Sprintf("Gray{creating output artefact tarball *%s*}", finalTarballName))
		tarballFileRef, err := os.Create(finalTarballName)
		if err != nil {
			return fmt.Errorf("failed to open output tarball: %w", err)
		}
		defer utils.SwallowClose(tarballFileRef)

		if err := builder.CreateArtefactTarball(os.DirFS(workDirectory), manifest, tarballFileRef); err != nil {
			return fmt.Errorf("failed to create artefact tarball: %w", err)
		}

		cmd.Println(bunt.Sprintf("LimeGreen{created artefact} *%s* LimeGreen{successfully!}", finalTarballName))

		return nil
	}

	return command
}

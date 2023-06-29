package cmd

import (
    "encoding/json"
    "fmt"
    "gitea.knockturnmc.com/marauder/client/pkg/builder"
    "gitea.knockturnmc.com/marauder/lib/pkg/artefact"
    "gitea.knockturnmc.com/marauder/lib/pkg/utils"
    "github.com/spf13/cobra"
    "os"
    "strings"
    "text/template"
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
    command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "p", "manifest.json", "location of the manifest file")
    command.PersistentFlags().StringVarP(&tarballName, "output", "o", "{{Identifier}}-{{Version}}-artefact.tar.gz", "name of the output tarball")

    command.RunE = func(cmd *cobra.Command, args []string) error {
        file, err := os.ReadFile(manifestFileLocation)
        if err != nil {
            return fmt.Errorf("failed to read %s: %w", manifestFileLocation, err)
        }

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

        tarballFileRef, err := os.Create(stringWriter.String())
        if err != nil {
            return fmt.Errorf("failed to open output tarball: %w", err)
        }
        defer utils.SwallowClose(tarballFileRef)

        artefactSourceDirectory := "."
        if len(args) > 0 {
            artefactSourceDirectory = args[0]
        }

        if err := builder.CreateArtefactTarball(os.DirFS(artefactSourceDirectory), manifest, tarballFileRef); err != nil {
            return fmt.Errorf("failed to create artefact tarball: %w", err)
        }

        return nil
    }

    return command
}

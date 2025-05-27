package cmd

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gonvenience/bunt"
	"github.com/knockturnmc/marauder/marauder-client/pkg/builder"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/filemodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

type OutputNameData struct {
	Identifier string
	Version    string
}

// BuildArtefactCommand constructs the command logic for the artefact creation.
func BuildArtefactCommand(
	configuration *Configuration,
) *cobra.Command {
	var (
		manifestFileLocation string
		tarballName          string
		sign                 bool
	)

	command := &cobra.Command{
		Use:   "artefact",
		Short: "Builds a marauder artefact into a tarball ready for publishing.",
		Args:  cobra.MaximumNArgs(1),
	}
	command.PersistentFlags().StringVarP(&manifestFileLocation, "manifest", "m", ".marauder.json", "location of the manifest file")
	command.PersistentFlags().StringVarP(&tarballName, "output", "o", "{{.Identifier}}-{{.Version}}-artefact.tar.gz", "name of the output tarball")
	command.PersistentFlags().BoolVar(&sign, "sign", true, "whether or not to sign the output artefact.")

	command.RunE = func(cmd *cobra.Command, args []string) error {
		workDirectory := "."
		if len(args) > 0 {
			workDirectory = args[0]
		}

		return buildArtefactInternalExecute(cmd, configuration, manifestFileLocation, tarballName, workDirectory, sign)
	}

	return command
}

// buildArtefactInternalExecute is the internal execution logic of the build artefact command.
func buildArtefactInternalExecute(
	cmd *cobra.Command,
	configuration *Configuration,
	manifestFileLocation, tarballName, workDirectory string,
	sign bool,
) error {
	// Parse the manifest.
	manifest, err := parseManifestFromDisk(cmd, manifestFileLocation, workDirectory)
	if err != nil {
		return err
	}

	// Parse the tarball name from the commandline flag
	finalTarballName, err := utils.ExecuteStringTemplateToString(tarballName, OutputNameData{
		Identifier: manifest.Identifier,
		Version:    manifest.Version,
	})
	if err != nil {
		return fmt.Errorf("failed to execute template for tarball output name: %w", err)
	}

	// Create the output file
	cmd.PrintErrln(bunt.Sprintf("Gray{creating output artefact tarball *%s*}", finalTarballName))
	tarballFileRef, err := os.Create(filepath.Clean(finalTarballName))
	if err != nil {
		return fmt.Errorf("failed to open output tarball: %w", err)
	}
	defer utils.SwallowClose(tarballFileRef)

	// Build and write the tarball to file.
	if err := builder.CreateArtefactTarball(os.DirFS(workDirectory), manifest, tarballFileRef); err != nil {
		return fmt.Errorf("failed to create artefact tarball: %w", err)
	}

	cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully build artefact}"))
	cmd.Println(finalTarballName)
	cmd.SetContext(context.WithValue(cmd.Context(), KeyBuildCommandTarballOutput, TarballBuildResult{
		TarballFileLocation:      finalTarballName,
		TarballSignatureLocation: "",
		Manifest:                 manifest,
	}))

	if sign {
		if err := signCreatedArtefact(cmd, configuration, tarballFileRef, finalTarballName); err != nil {
			return err
		}
	}

	return nil
}

// parseManifestFromDisk parses the manifest from the disk with the given name in the given work directory.
func parseManifestFromDisk(cmd *cobra.Command, manifestFileLocation string, workDirectory string) (filemodel.Manifest, error) {
	file, err := os.ReadFile(filepath.Clean(manifestFileLocation))
	if err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed to read %s: %w", manifestFileLocation, err)
	}

	var manifest filemodel.Manifest

	cmd.PrintErrln(bunt.Sprintf("Gray{fetching build information from project}"))
	buildInformation, err := builder.FetchBuildInformation(workDirectory)
	if err != nil {
		cmd.PrintErrln(bunt.Sprintf("Red{failed to parse build information, excluding them: %s}", err.Error()))
		timestamp := time.Now()
		buildInformation = filemodel.BuildInformation{
			Repository:           "nan",
			Branch:               "nan",
			CommitUser:           "nan",
			CommitEmail:          "nan",
			CommitHash:           "nan",
			CommitMessage:        "nan",
			Timestamp:            timestamp,
			BuildSpecificVersion: "t" + strconv.FormatInt(timestamp.Unix(), 10),
		}
	} else {
		manifest.BuildInformation = &buildInformation
	}

	// Parse the manifest file
	cmd.PrintErrln(bunt.Sprintf("Gray{parsing manifest file %s}", manifestFileLocation))

	templatedManifestContent, err := utils.ExecuteStringTemplateToString(string(file), struct {
		Build filemodel.BuildInformation
	}{
		Build: buildInformation,
	})
	if err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed to resolve templates in manifest file: %w", err)
	}

	if err := json.Unmarshal([]byte(templatedManifestContent), &manifest); err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest, nil
}

// signCreatedArtefact signs the tarball file ref and stores it under the same name .sig.
func signCreatedArtefact(cmd *cobra.Command, configuration *Configuration, tarballFileRef *os.File, tarballName string) error {
	key, err := configuration.ParseSigningKey()
	if err != nil {
		return fmt.Errorf("failed to parse signing key: %w", err)
	}

	if _, err := tarballFileRef.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek 0 index on file for signature computation: %w", err)
	}

	sha256, err := utils.ComputeSha256(tarballFileRef)
	if err != nil {
		return fmt.Errorf("failed to compute signature: %w", err)
	}

	signature, err := key.Sign(rand.Reader, sha256)
	if err != nil {
		return fmt.Errorf("failed to sign sha hashsum: %w", err)
	}

	signatureFileName := tarballName + ".sig"
	if err := os.WriteFile(signatureFileName, ssh.Marshal(signature), 0o600); err != nil {
		return fmt.Errorf("failed to write signature to %s: %w", signatureFileName, err)
	}

	cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully signed artefact}"))
	cmd.Println(signatureFileName)

	result, ok := cmd.Context().Value(KeyBuildCommandTarballOutput).(TarballBuildResult)
	if !ok {
		return fmt.Errorf("failed to retrieve existing context: %w", ErrContextMissingValue)
	}

	result.TarballSignatureLocation = signatureFileName
	cmd.SetContext(context.WithValue(cmd.Context(), KeyBuildCommandTarballOutput, result))

	return nil
}

// TarballBuildResult is a simple helper struct stored in the context of the build command to indicate the location of the output.
type TarballBuildResult struct {
	Manifest                 filemodel.Manifest
	TarballFileLocation      string
	TarballSignatureLocation string
}

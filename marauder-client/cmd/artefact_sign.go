package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"path"

	"github.com/gonvenience/bunt"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func ArtefactSignCommand(configuration *Configuration) *cobra.Command {
	var outputFileTemplate string
	command := &cobra.Command{
		Use:   "sign",
		Short: "Allows marauder to sign artefacts for controller",
		Args:  cobra.ExactArgs(1),
	}

	command.PersistentFlags().StringVarP(
		&outputFileTemplate, "outputName", "o", "{{.File}}.sig",
		"the name of the output file relative to the signed file",
	)

	command.RunE = func(cmd *cobra.Command, args []string) error {
		fileToSignPath := args[0]
		outputFilePath, err := utils.ExecuteStringTemplateToString(outputFileTemplate, struct{ File string }{File: path.Base(fileToSignPath)})
		if err != nil {
			return fmt.Errorf("failed to compute output file path: %w", err)
		}

		key, err := configuration.ParseSigningKey()
		if err != nil {
			return fmt.Errorf("failed to parse signing key: %w", err)
		}

		fileToSign, err := os.Open(fileToSignPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s to sign: %w", fileToSignPath, err)
		}

		defer func() { _ = fileToSign.Close() }()
		sha256, err := utils.ComputeSha256(fileToSign)
		if err != nil {
			return fmt.Errorf("failed to compute sha256 hash for file to sign: %w", err)
		}

		sign, err := key.Sign(rand.Reader, sha256)
		if err != nil {
			return fmt.Errorf("failed to sign file hash: %w", err)
		}

		if err := os.WriteFile(outputFilePath, ssh.Marshal(sign), 0o600); err != nil {
			return fmt.Errorf("failed to write signature to %s: %w", outputFilePath, err)
		}

		cmd.PrintErrln(bunt.Sprintf("LimeGreen{successfully signed artefact}"))
		cmd.Println(outputFilePath)

		return nil
	}

	return command
}

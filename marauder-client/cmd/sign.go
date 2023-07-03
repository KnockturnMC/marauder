package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/user"
	"path"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func SignCommand() *cobra.Command {
	var (
		privateKeyFilePath string
		outputFileTemplate string
	)
	command := &cobra.Command{
		Use:   "sign",
		Short: "Allows marauder to sign files for controller",
		Args:  cobra.ExactArgs(1),
	}
	command.PersistentFlags().StringVarP(
		&privateKeyFilePath, "privateKey", "p", "{{.User.HomeDir}}/.config/marauder/key",
		"the private key file used for signing the file.",
	)
	_ = command.MarkFlagRequired("privateKey")
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

		userAccount, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to fetch current user: %w", err)
		}
		privateKeyFilePath, err := utils.ExecuteStringTemplateToString(privateKeyFilePath, struct{ User *user.User }{User: userAccount})
		if err != nil {
			return fmt.Errorf("failed to evaluate private key file path: %w", err)
		}

		privateKeyBytes, err := os.ReadFile(privateKeyFilePath)
		if err != nil {
			return fmt.Errorf("failed to read private key file for signing: %w", err)
		}

		key, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key bytes: %w", err)
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

		return nil
	}

	return command
}

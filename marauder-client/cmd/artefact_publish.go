package cmd

import (
    "bytes"
    "fmt"
    "gitea.knockturnmc.com/marauder/lib/pkg/utils"
    "github.com/gonvenience/bunt"
    "github.com/spf13/cobra"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "os/user"
)

func ArtefactPublishCommand() *cobra.Command {
    var (
        tlsPath                   string
        artefactFilePath          string
        artefactFileSignaturePath string
    )

    command := &cobra.Command{
        Use:   "publish",
        Short: "Publishes a marauder artefact to a controller",
        Args:  cobra.ExactArgs(1),
    }
    command.PersistentFlags().StringVar(
        &tlsPath, "tls", "{{.User.HomeDir}}/.config/marauder/tls",
        "the root folder for the tls file expected by marauder, specifically a cert.pem and a key.pem.",
    )
    command.PersistentFlags().StringVarP(
        &artefactFilePath, "artefactFile", "f", "",
        "the name of the output file relative to the signed file",
    )
    command.PersistentFlags().StringVarP(
        &artefactFilePath, "artefactSignature", "s", "",
        "the name of the output file relative to the signed file",
    )

    _ = command.MarkPersistentFlagRequired("artefactFile")

    command.RunE = func(cmd *cobra.Command, args []string) error {
        // default signature path
        if artefactFileSignaturePath == "" {
            artefactFileSignaturePath = artefactFilePath + ".sig"
        }

        userAccount, err := user.Current()
        if err != nil {
            return fmt.Errorf("failed to fetch current user: %w", err)
        }
        tlsPath, err = utils.ExecuteStringTemplateToString(tlsPath, struct{ User *user.User }{User: userAccount})
        if err != nil {
            return fmt.Errorf("failed to evaluate tls file path: %w", err)
        }

        controllerHost := args[0]
        httpClient, err := httpClientWithPotentialTLS(tlsPath)
        if err != nil {
            cmd.Println(bunt.Sprintf("Red{failed to enable tls: %s}", err))
        }

        artefactFile, err := os.Open(artefactFilePath)
        if err != nil {
            return fmt.Errorf("failed to open artefact file %s: %w", artefactFilePath, err)
        }

        defer func() { _ = artefactFile.Close() }()

        artefactSignature, err := os.Open(artefactFileSignaturePath)
        if err != nil {
            return fmt.Errorf("failed to open artefact signature file %s: %w", artefactFileSignaturePath, err)
        }

        defer func() { _ = artefactSignature.Close() }()

        var body bytes.Buffer
        multipartWriter := multipart.NewWriter(&body)

        // Write artefact
        artefactFileUpload, err := multipartWriter.CreateFormFile("upload", "artefact")
        if err != nil {
            return fmt.Errorf("failed to create form file for artefact: %w", err)
        }

        if _, err := io.Copy(artefactFileUpload, artefactFile); err != nil {
            return fmt.Errorf("failed to write artefaact to form header: %w", err)
        }

        // write signature
        artefactSigUpload, err := multipartWriter.CreateFormFile("upload", "signature")
        if err != nil {
            return fmt.Errorf("failed to create form file for artefact signature: %w", err)
        }

        if _, err := io.Copy(artefactSigUpload, artefactSignature); err != nil {
            return fmt.Errorf("failed to write artefaact signature to form header: %w", err)
        }

        response, err := httpClient.Post(controllerHost, multipartWriter.FormDataContentType(), &body)
        if err != nil {
            return fmt.Errorf("failed to execute post request: %w", err)
        }

        if response.StatusCode >= 300 {
            return nil
        }

        cmd.Println(bunt.Sprintf("LimeGreen{Uploaded artefact %s to controller!}", artefactFilePath))
        return nil
    }

    return command
}

// httpClientWithPotentialTLS creates a new http client given the cert.pem and key.pem files.
// This method will always return a usable http client, potentially with an error if no tls could be configured.
func httpClientWithPotentialTLS(tlsPath string) (*http.Client, error) {
    configuration, err := utils.ParseTLSConfiguration(tlsPath)
    if err != nil {
        return http.DefaultClient, err
    }

    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: configuration,
        },
    }, nil
}

package utils

import (
	"bytes"
	"fmt"
	"os/user"
	"text/template"
)

// ExecuteStringTemplateToString executes the template string as a template given the passed data and returns the result
// as a string.
func ExecuteStringTemplateToString[Data any](templateString string, data Data) (string, error) {
	parsed, err := template.New("").Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buffer bytes.Buffer
	if err := parsed.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buffer.String(), nil
}

// EvaluateFilePathTemplate evaluates a file path template for a cobra command.
func EvaluateFilePathTemplate(filePathTemplate string) (string, error) {
	// fetch user and expand tls path
	// If we cannot fully fetch the user, the env does not support user fetching. Proceed anyway, they could not use variables
	// in their path. They get a nil user, if they access it the templating will fail.
	userAccount, _ := user.Current()

	filePathTemplate, err := ExecuteStringTemplateToString(filePathTemplate, struct{ User *user.User }{User: userAccount})
	if err != nil {
		return "", fmt.Errorf("failed to evaluate filePathTemplate: %w", err)
	}

	return filePathTemplate, nil
}

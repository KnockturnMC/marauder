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
	userAccount, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to fetch current user: %w", err)
	}
	filePathTemplate, err = ExecuteStringTemplateToString(filePathTemplate, struct{ User *user.User }{User: userAccount})
	if err != nil {
		return "", fmt.Errorf("failed to evaluate filePathTemplate: %w", err)
	}

	return filePathTemplate, nil
}

package utils

import (
	"bytes"
	"fmt"
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

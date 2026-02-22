package middleware

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gonvenience/bunt"
	"github.com/sirupsen/logrus"
)

// The GinLikeJSONFormatter is a logger similar to the gin logger that prints errors to the console.
type GinLikeJSONFormatter struct{}

func (g GinLikeJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	g.sanitizeDataFields(entry)

	fieldsAsJSON, fieldToJSONErr := json.Marshal(entry.Data)
	if fieldToJSONErr != nil {
		return nil, fmt.Errorf("failed to marshal entry fields to json: %w", fieldToJSONErr)
	}

	callerFile := "nil"
	callerLine := "nil"
	if entry.Caller != nil {
		callerFile = entry.Caller.File
		callerLine = strconv.Itoa(entry.Caller.Line)
	}

	return fmt.Appendf([]byte{}, "[marauder] %v |%s| %s | %s:%s |  %s\n",
		entry.Time.Format("2006/01/02 - 15:04:05"),
		bunt.Sprintf(buntColorNameForEntry(entry)+"{"+entry.Level.String()+"}"),
		entry.Message,
		callerFile, callerLine,
		fieldsAsJSON,
	), nil
}

// SanitizeDataFields is responsible for sanitizing data fields in an entry to ensure
// they are properly marshaled into the final log.
func (g GinLikeJSONFormatter) sanitizeDataFields(entry *logrus.Entry) {
	for key, value := range entry.Data {
		switch typedVal := value.(type) {
		case error:
			entry.Data[key] = g.sanitizeDataFieldError(typedVal)
		default:
		}
	}
}

// sanitizeDataFieldError is responsible for sanitizing an error into a string representation
// for a logrus data field entry.
func (g GinLikeJSONFormatter) sanitizeDataFieldError(err error) string {
	return err.Error()
}

// The buntColorNameForEntry attempts to find a fitting color for the entry.
func buntColorNameForEntry(entry *logrus.Entry) string {
	switch entry.Level {
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return "Red"
	case logrus.WarnLevel:
		return "Yellow"
	case logrus.InfoLevel:
		return "Green"
	case logrus.DebugLevel:
		return "Cyan"
	case logrus.TraceLevel:
		return "Gray"
	}

	return "Cyan"
}

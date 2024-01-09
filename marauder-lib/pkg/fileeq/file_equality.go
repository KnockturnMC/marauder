package fileeq

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"gopkg.in/yaml.v3"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

// ErrUnknownFileEquality may be returned if a file equality is not found for a string identifier in a FileEqualityRegistry.
var ErrUnknownFileEquality = errors.New("unknown file equality")

// The FileEqualityRegistry allows map-like access to file equality implementations for later consumption.
type FileEqualityRegistry map[string]FileEquality

// DefaultFileEqualityRegistry constructs a new, default filled, file equality registry.
func DefaultFileEqualityRegistry() FileEqualityRegistry {
	return FileEqualityRegistry{
		"hash": Sha256SumFileEquality{},
		"noop": NOOPFileEquality{},
		"json": JSONFileEquality{},
		"yaml": YAMLFileEquality{},
	}
}

// The FileEquality interface defines a single function that determines if two files, based on their content
// are to be considered equal.
type FileEquality interface {
	// Equals determines if the two files, represented by their reader, are equal.
	// The method will completely read both readers but will not close them.
	// If an error occurs while reading either of them, an error is returned, meaning the comparison
	// was not capable of being computed.
	// This is inherently different from a comparison that yielded false.
	Equals(first io.Reader, second io.Reader) (bool, error)
}

// The Sha256SumFileEquality compares two files via their sha256hash.
type Sha256SumFileEquality struct{}

func (s Sha256SumFileEquality) Equals(first io.Reader, second io.Reader) (bool, error) {
	return compareReadersBy(first, second, func(reader io.Reader) (string, error) {
		hash, err := utils.ComputeSha256(reader)
		if err != nil {
			return "", fmt.Errorf("failed to compute hash for reader: %w", err)
		}

		return hex.EncodeToString(hash), nil
	})
}

// The NOOPFileEquality considers any two files equal.
type NOOPFileEquality struct{}

func (n NOOPFileEquality) Equals(_ io.Reader, _ io.Reader) (bool, error) {
	return true, nil
}

// The JSONFileEquality compares files by parsing them as json.
// If either file cannot be parsed as json comparison errors (not fails).
type JSONFileEquality struct{}

func (j JSONFileEquality) Equals(first io.Reader, second io.Reader) (bool, error) {
	return compareReadersByCmp(first, second, func(reader io.Reader) (interface{}, error) {
		var result interface{}
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode json: %w", err)
		}

		return result, nil
	}, reflect.DeepEqual)
}

// The YAMLFileEquality compares files by parsing them as yaml.
// If either file cannot be parsed as yaml comparison errors (not fails).
type YAMLFileEquality struct{}

func (y YAMLFileEquality) Equals(first io.Reader, second io.Reader) (bool, error) {
	return compareReadersByCmp(first, second, func(reader io.Reader) (interface{}, error) {
		var result interface{}
		if err := yaml.NewDecoder(reader).Decode(&result); err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("failed to decode yaml: %w", err)
		}

		return result, nil
	}, reflect.DeepEqual)
}

// compareReadersBy is a utility method that compares two readers by mapping them via the passed mapper function and comparing their result.
func compareReadersBy[T comparable](first io.Reader, second io.Reader, mapper func(reader io.Reader) (T, error)) (bool, error) {
	return compareReadersByCmp(first, second, mapper, func(first T, second T) bool {
		return first == second
	})
}

// compareReadersBy is a utility method that compares two readers by mapping them via the passed mapper function and comparing their result.
func compareReadersByCmp[T any](
	first io.Reader,
	second io.Reader,
	mapper func(reader io.Reader) (T, error),
	cmp func(first T, second T) bool,
) (bool, error) {
	firstMapped, err := mapper(first)
	if err != nil {
		return false, fmt.Errorf("failed to map first reader for comparison: %w", err)
	}

	secondMapped, err := mapper(second)
	if err != nil {
		return false, fmt.Errorf("failed to map second reader for comparison: %w", err)
	}

	return cmp(firstMapped, secondMapped), nil
}

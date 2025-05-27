package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/knockturnmc/marauder/marauder-lib/pkg"
)

// The TLSConfiguration represents a tls configuration for any marauder application.
// It may either point to a specific path or provide the entire paths to each relevant file.
type TLSConfiguration struct {
	Folder *string                `yaml:"folder,omitempty"`
	Files  *TLSConfigurationFiles `yaml:"files,omitempty"`
}

// The TLSConfigurationFiles struct holds the specific references to the tls configuration files.
type TLSConfigurationFiles struct {
	Key             string `yaml:"key"`
	Certificate     string `yaml:"certificate"`
	RootCertificate string `yaml:"rootCertificate"`
}

// ErrNoTLSPathConfigured is returned if no tls path is configured.
var ErrNoTLSPathConfigured = errors.New("no tls path configured")

// ParseTLSConfigurationFromType parses the tls configuration from the provided tls yaml config.
func ParseTLSConfigurationFromType(configuration TLSConfiguration) (*tls.Config, error) {
	switch {
	case configuration.Files != nil:
		certificatePath, err := EvaluateFilePathTemplate(configuration.Files.Certificate)
		if err != nil {
			return nil, fmt.Errorf("failed to expand certificate path: %w", err)
		}

		keyPath, err := EvaluateFilePathTemplate(configuration.Files.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to expand key path: %w", err)
		}

		rootCertPath, err := EvaluateFilePathTemplate(configuration.Files.RootCertificate)
		if err != nil {
			return nil, fmt.Errorf("failed to expand root cert path: %w", err)
		}

		return ParseTLSConfigurationFromPaths(certificatePath, keyPath, rootCertPath)
	case configuration.Folder != nil:
		topFolderPath, err := EvaluateFilePathTemplate(*configuration.Folder)
		if err != nil {
			return nil, fmt.Errorf("failed to expand top folder path: %w", err)
		}

		return ParseTLSConfigurationTopPath(topFolderPath)
	default:
		return nil, fmt.Errorf("config did not contain folder or files: %w", ErrNoTLSPathConfigured)
	}
}

// ParseTLSConfigurationTopPath parses a bare bone tls configuration from a folder.
// Specifically, three files are parsed.
// pool/ the pool of intermediate certs
// tls.crt - the configurations own certificate.
// tls.key - the configurations own public key.
func ParseTLSConfigurationTopPath(tlsPath string) (*tls.Config, error) {
	if tlsPath == "" {
		return nil, ErrNoTLSPathConfigured
	}

	certPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSCertificateFileName)
	keyPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSKeyFileName)
	poolPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSPoolDir)
	poolPathFiles, err := os.ReadDir(filepath.Clean(poolPath))
	if err != nil {
		return nil, fmt.Errorf("failed to list pool path files: %w", err)
	}

	rootCerts := make([]string, 0)

	for _, file := range poolPathFiles {
		if file.IsDir() {
			continue
		}

		rootCerts = append(rootCerts, filepath.Join(poolPath, file.Name()))
	}

	return ParseTLSConfigurationFromPaths(certPath, keyPath, rootCerts...)
}

// ParseTLSConfigurationFromPaths parses a bare bone tls configuration from the specific set of files.
// This expects the paths to be fully resolved.
func ParseTLSConfigurationFromPaths(certPath string, keyPath string, rootCerts ...string) (*tls.Config, error) {
	certificateBytes, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load our certificate at %s: %w", certPath, err)
	}

	key, err := os.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load key at %s: %w", keyPath, err)
	}

	certificate, err := tls.X509KeyPair(certificateBytes, key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 key pair: %w", err)
	}

	certPool := x509.NewCertPool()
	for _, certificatePath := range rootCerts {
		poolEntryBytes, err := os.ReadFile(filepath.Clean(certificatePath))
		if err != nil {
			return nil, fmt.Errorf("failed to read cert %s: %w", certificatePath, err)
		}

		certPool.AppendCertsFromPEM(poolEntryBytes)
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

package utils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gitea.knockturnmc.com/marauder/lib/pkg"
)

// ErrNoTLSPathConfigured is returned if no tls path is configured.
var ErrNoTLSPathConfigured = errors.New("no tls path configured")

// ParseTLSConfiguration parses a bare bone tls configuration from a folder.
// Specifically, three files are parsed.
// pool/ the pool of intermediate certs
// tls.crt - the configurations own certificate.
// tls.key - the configurations own public key.
func ParseTLSConfiguration(tlsPath string) (*tls.Config, error) {
	if tlsPath == "" {
		return nil, ErrNoTLSPathConfigured
	}

	certPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSCertificateFileName)
	keyPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSKeyFileName)
	poolPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSPoolDir)
	certificateBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load our certificate at %s: %w", certPath, err)
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load key at %s: %w", keyPath, err)
	}

	certificate, err := tls.X509KeyPair(certificateBytes, key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 key pair: %w", err)
	}

	certPool := x509.NewCertPool()
	err = filepath.WalkDir(poolPath, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		poolEntryBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		certPool.AppendCertsFromPEM(poolEntryBytes)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load chain certificates: %w", err)
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

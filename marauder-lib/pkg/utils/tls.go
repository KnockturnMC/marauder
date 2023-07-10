package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gitea.knockturnmc.com/marauder/lib/pkg"
	"os"
	"path/filepath"
)

// ParseTLSConfiguration parses a bare bone tls configuration from a folder.
// Specifically, three files are parsed.
// root.ca - the root certificate
// cert.pem - the configurations own certificate.
// key.pem - the configurations own public key.
func ParseTLSConfiguration(tlsPath string) (*tls.Config, error) {
	certPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSCertificateFileName)
	keyPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSKeyFileName)
	rootCertPath := fmt.Sprintf("%s%c%s", tlsPath, filepath.Separator, pkg.TLSRootCertificate)
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

	rootCertBytes, err := os.ReadFile(rootCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read root certificate %s: %w", rootCertBytes, err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(rootCertBytes)

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{certificate},
	}, nil
}

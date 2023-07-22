package keyauth

import (
	"bufio"
	"fmt"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"golang.org/x/crypto/ssh"
)

// ParseKnownPublicKeys parses a list of public keys from an ssh-like authorized_keys file.
// The authorizedKeyPath is expanded using utils.EvaluateFilePathTemplate.
func ParseKnownPublicKeys(authorizedKeyPath string) ([]ssh.PublicKey, error) {
	authorizedKeyPath, err := utils.EvaluateFilePathTemplate(authorizedKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorized key path: %w", err)
	}

	// Parse authorized keys from disk
	authorizedKeys := make([]ssh.PublicKey, 0)
	authorizedKeysFile, err := os.Open(authorizedKeyPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open authorized key file: %w", err)
		}

		return authorizedKeys, nil
	}

	defer func() { _ = authorizedKeysFile.Close() }()

	scanner := bufio.NewScanner(authorizedKeysFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		out, _, _, _, err := ssh.ParseAuthorizedKey(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to parse authorizsed key %s: %w", scanner.Text(), err)
		}

		authorizedKeys = append(authorizedKeys, out)
	}

	return authorizedKeys, nil
}

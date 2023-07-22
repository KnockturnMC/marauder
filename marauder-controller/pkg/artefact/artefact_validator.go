package artefact

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"gitea.knockturnmc.com/marauder/lib/pkg/worker"
	"golang.org/x/crypto/ssh"
)

var (
	// ErrUnknownSignature is returned if a signature on an artefact is unknown to the validator.
	ErrUnknownSignature = errors.New("unknown signature")

	// ErrHashMismatch is returned if the hashes of a manifest do not match the content of the artefact.
	ErrHashMismatch = errors.New("mismatching hashes")
)

// The ValidationResult is returned by the Validator via a channel once the validation is completed.
type ValidationResult struct {
	Manifest     filemodel.Manifest
	ArtefactHash []byte
}

// The Validator is a worker queue responsible for validating a newly uploaded artefact.
type Validator interface {
	// SubmitArtefact submits the artefact tarball located at the given path to the validator.
	// This method is non-blocking and will only submit the path.
	// The actual validation may happen at any point afterward.
	SubmitArtefact(artefactPath, signaturePath string) <-chan worker.Outcome[ValidationResult]
}

// WorkedBasedValidator represents a artefact validator based on a worker.Dispatcher instance.
type WorkedBasedValidator struct {
	dispatcher      *worker.Dispatcher[ValidationResult]
	knownPublicKeys []ssh.PublicKey
}

func NewWorkedBasedValidator(dispatcher *worker.Dispatcher[ValidationResult], knownPublicKeys []ssh.PublicKey) *WorkedBasedValidator {
	return &WorkedBasedValidator{dispatcher: dispatcher, knownPublicKeys: knownPublicKeys}
}

// SubmitArtefact submits the artefact and its signature to the validator.
func (w *WorkedBasedValidator) SubmitArtefact(artefactPath, signaturePath string) <-chan worker.Outcome[ValidationResult] {
	return w.dispatcher.Dispatch(func() (ValidationResult, error) {
		signature, err := os.ReadFile(signaturePath)
		if err != nil {
			return ValidationResult{}, fmt.Errorf("failed to read signature %s: %w", signaturePath, err)
		}

		artefactFile, err := os.Open(artefactPath)
		if err != nil {
			return ValidationResult{}, fmt.Errorf("failed to artefactFile artefact tarball: %w", err)
		}

		defer func() { _ = artefactFile.Close() }()

		artefactHash, err := w.verifyArtefactSignature(artefactFile, signature)
		if err != nil {
			return ValidationResult{}, fmt.Errorf("failed to verify artefact signature: %w", err)
		}

		manifest, err := w.verifyArtefactManifestHashes(artefactFile)
		if err != nil {
			return ValidationResult{}, fmt.Errorf("failed to verify hashes of artefact: %w", err)
		}

		return ValidationResult{
			Manifest:     manifest,
			ArtefactHash: artefactHash,
		}, nil
	})
}

// verifyArtefactSignature verifies the uploaded signature against the artefact file by checking if the signature is
// a) valid for the artefact file.
// b) belongs to a known public key of marauderctl.
func (w *WorkedBasedValidator) verifyArtefactSignature(artefact *os.File, signatureBytes []byte) ([]byte, error) {
	if _, err := artefact.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset artefact file ref to start: %w", err)
	}

	var signature ssh.Signature
	if err := ssh.Unmarshal(signatureBytes, &signature); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature bytes: %w", err)
	}

	sha256, err := utils.ComputeSha256(artefact)
	if err != nil {
		return nil, fmt.Errorf("failed to compute sha256 hash for artefact tarball: %w", err)
	}

	for _, key := range w.knownPublicKeys {
		if err := key.Verify(sha256, &signature); err != nil {
			continue
		}

		return sha256, nil // Return null if a key verified.
	}

	return nil, fmt.Errorf("did not find signature in %d known signatures: %w", len(w.knownPublicKeys), ErrUnknownSignature)
}

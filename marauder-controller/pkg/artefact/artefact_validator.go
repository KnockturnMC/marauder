package artefact

// The ValidationResult is returned by the Validator via a channel once the validation is completed.
type ValidationResult struct {
	// Err holds an potential error that occurred during the validation.
	Err error
}

// The Validator is a worker queue responsible for validating a newly uploaded artefact.
type Validator interface {
	// SubmitArtefact submits the artefact tarball located at the given path to the validator.
	// This method is non-blocking and will only submit the path.
	// The actual validation may happen at any point afterward.
	SubmitArtefact(path string) <-chan ValidationResult
}

package filemodel

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrMaxFileMatchesFailed is returned if a matcher/builder for the file manifest matches more than the restricted maximum.
	ErrMaxFileMatchesFailed = errors.New("more than max amount matched")

	// ErrMinFileMatchesFailed is returned if a matcher/builder for the file manifest matches less than the restricted minimum.
	ErrMinFileMatchesFailed = errors.New("less than minimal amount matched")

	// ErrExactFileMatchesFailed is returned if a matcher/builder for the file manifest matches a different amount than the exact match count.
	ErrExactFileMatchesFailed = errors.New("different than exact amount matched")
)

// The Hashes type holds the hashes of a collections of files.
type Hashes map[string]string

// The Manifest type defines an artefact's manifest managed by marauder.
type Manifest struct {
	// The unique, marauder wide Identifier of the artefact, usually the name of the plugin the artefact is created for.
	// An example would be `spellcore`.
	Identifier string `json:"identifier"`

	// The Version of the artefact, following schematic versioning.
	Version string `json:"version"`

	// The Files included in this artefact, not flattened.
	// The file reference may hence include specific files or references to whole folders in the artefact.
	Files []FileReference `json:"files"`

	// BuildInformation holds additional information about the manifest based on a potential build.
	// This field is optional as artefacts might be constructed without build information attached.
	BuildInformation *BuildInformation `json:"buildInformation,omitempty"`

	// DeploymentTargets define into which server/environment an artefact can generally be deployed into
	// Not all servers require the deployment of a specific artefact, hence this field actively defines which servers
	// should be targeted during a release.
	DeploymentTargets DeploymentTargets `json:"deploymentTargets,omitempty"`

	// Hashes contains a collection of hashes for each fully resolved file in the manifest.
	// While the Files field may hold the globs and targets of specific files, this
	// field holds a full list of all included files with their hashes.
	// This cannot be archived on a folder level, as deployments might deploy into folders holding other data.
	Hashes Hashes `json:"hashes,omitempty"`
}

// The BuildInformation struct holds potential additional information about the build the manifest was generated for.
type BuildInformation struct {
	// Repository represents a reference to the git repository this build originated from.
	Repository string `json:"repository"`

	// Branch defines the branch the build originated from.
	Branch string `json:"branch"`

	// CommitUser holds the user.name of the committer authoring the commit this build was produced from.
	CommitUser string `json:"commitUser"`

	// CommitEmail provides the email of the committer authoring the commit this build was produced from.
	CommitEmail string `json:"commitEmail"`

	// CommitUser holds the hash of the commit this build was produced from.
	CommitHash string `json:"commitHash"`

	// CommitMessage provides the full message of the commit this build was produced from.
	CommitMessage string `json:"commitMessage"`

	// Timestamp represents the time at which the build information were gathered.
	Timestamp time.Time `json:"timestamp"`

	// The BuildSpecificVersion represents a unique version string generated specifically for this build.
	// This value may be based on either the CommitHash or the Timestamp.
	BuildSpecificVersion string `json:"buildSpecificVersion"`
}

// FileReference defines a specific configuration of an artefacts file as defined by its manifest.
type FileReference struct {
	// The target path as found in the artefacts file collection as well as its final location on the server.
	Target string `json:"target"`

	// A string representation of a glob that identifies the files during the ci build of the project that produces the artefact.
	CISourceGlob string `json:"ciSourceGlob"`

	// A restriction type that may be used to restrict what/how files are matched.
	Restrictions *FileRestriction `json:"restrictions,omitempty"`
}

// The FileRestriction type allows to restrict matches by marauder during the artefact building process.
type FileRestriction struct {
	// Max defines how many files can be matched at max.
	// If more files are matched, marauder will error before building.
	Max *int `json:"max,omitempty"`

	// Min defines how many files can be matched at min.
	// If less files are matched, marauder will error before building.
	Min *int `json:"min,omitempty"`

	// Exact defines how many files can be matched exactly.
	// If any different amount of files were matched, marauder will error.
	Exact *int `json:"exact,omitempty"`
}

// ValidateMatchAmount validates if the passed amount matches the file restrictions.
func (f *FileRestriction) ValidateMatchAmount(amountMatched int) error {
	if f == nil {
		return nil
	}

	switch {
	case f.Exact != nil && *f.Exact != amountMatched:
		return fmt.Errorf("matched %d (expected %d) files: %w", amountMatched, *f.Exact, ErrExactFileMatchesFailed)
	case f.Min != nil && *f.Min > amountMatched:
		return fmt.Errorf("matched %d (expected > %d) files: %w", amountMatched, *f.Min, ErrMinFileMatchesFailed)
	case f.Max != nil && *f.Max < amountMatched:
		return fmt.Errorf("matched %d (expected < %d) files: %w", amountMatched, *f.Max, ErrMaxFileMatchesFailed)
	}

	return nil
}

// The DeploymentTargets type holds a map of environments to a slice of servers in said environment that the artefact should be deployed to.
type DeploymentTargets map[string][]string

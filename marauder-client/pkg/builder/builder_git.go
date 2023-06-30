package builder

import (
	"fmt"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"github.com/go-git/go-git/v5"
)

// InsertBuildInformation fetches and inserts build information from the current environment.
func InsertBuildInformation(path string, manifest artefact.Manifest) (artefact.Manifest, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return artefact.Manifest{}, fmt.Errorf("failed to open git repository at %s: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return artefact.Manifest{}, fmt.Errorf("failed to retrieve head of repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return artefact.Manifest{}, fmt.Errorf("failed to retrieve origin remote: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return artefact.Manifest{}, fmt.Errorf("failed to locate commit information from head hash: %w", err)
	}

	information := artefact.BuildInformation{
		Repository:    remote.Config().URLs[0],
		Branch:        strings.Replace(head.Name().String(), "refs/head/", "", 1),
		CommitUser:    commit.Author.Name,
		CommitEmail:   commit.Author.Email,
		CommitHash:    commit.Hash.String(),
		CommitMessage: commit.Message,
	}

	manifest.BuildInformation = &information

	return manifest, nil
}

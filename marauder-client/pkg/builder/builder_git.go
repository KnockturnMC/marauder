package builder

import (
	"fmt"
	"strings"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"github.com/go-git/go-git/v5"
)

// FetchBuildInformation fetches and inserts build information from the current environment.
func FetchBuildInformation(path string) (artefact.BuildInformation, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return artefact.BuildInformation{}, fmt.Errorf("failed to open git repository at %s: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return artefact.BuildInformation{}, fmt.Errorf("failed to retrieve head of repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return artefact.BuildInformation{}, fmt.Errorf("failed to retrieve origin remote: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return artefact.BuildInformation{}, fmt.Errorf("failed to locate commit information from head hash: %w", err)
	}

	information := artefact.BuildInformation{
		Repository:           remote.Config().URLs[0],
		Branch:               strings.Replace(head.Name().String(), "refs/head/", "", 1),
		CommitUser:           commit.Author.Name,
		CommitEmail:          commit.Author.Email,
		CommitHash:           commit.Hash.String(),
		CommitMessage:        commit.Message,
		Timestamp:            time.Now(),
		BuildSpecificVersion: "g" + commit.Hash.String()[:7],
	}

	return information, nil
}

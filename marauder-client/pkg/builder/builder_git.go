package builder

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"github.com/go-git/go-git/v5"
)

// FetchBuildInformation fetches and inserts build information from the current environment.
func FetchBuildInformation(path string) (filemodel.BuildInformation, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return filemodel.BuildInformation{}, fmt.Errorf("failed to open git repository at %s: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return filemodel.BuildInformation{}, fmt.Errorf("failed to retrieve head of repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return filemodel.BuildInformation{}, fmt.Errorf("failed to retrieve origin remote: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return filemodel.BuildInformation{}, fmt.Errorf("failed to locate commit information from head hash: %w", err)
	}

	// Allow overriding the branch name fetched if e.g. this isn't run in a fully checked out repo.
	branchFromGitRef := strings.Replace(head.Name().String(), "refs/head/", "", 1)
	branchFromEnv, found := os.LookupEnv(pkg.MarauderEnvironmentBranchOverride)
	if found {
		branchFromGitRef = branchFromEnv
	}

	// Allow overriding the build specific version fetched if e.g. this isn't run in a fully checked out repo.
	buildSpecificVersionFromGit := "g" + commit.Hash.String()[:7]
	buildSpecificVersionFromEnv, found := os.LookupEnv(pkg.MarauderEnvironmentBuildSpecificVersionOverride)
	if found {
		buildSpecificVersionFromGit = buildSpecificVersionFromEnv
	}

	information := filemodel.BuildInformation{
		Repository:           remote.Config().URLs[0],
		Branch:               branchFromGitRef,
		CommitUser:           commit.Author.Name,
		CommitEmail:          commit.Author.Email,
		CommitHash:           commit.Hash.String(),
		CommitMessage:        strings.TrimRight(commit.Message, "\n"),
		Timestamp:            time.Now(),
		BuildSpecificVersion: buildSpecificVersionFromGit,
	}

	return information, nil
}

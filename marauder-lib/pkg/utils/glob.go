package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

// ErrFailedToFindShortestPath is returned by FindShortestMatch if the file does not match the glob.
var ErrFailedToFindShortestPath = errors.New("files does not match glob")

// ShortestGlobPathCache holds a map of globs to a list of existing shortest matches.
type ShortestGlobPathCache struct {
	Cache map[string][]string
}

// NewShortestGlobPathCache creates a new cache for computing the shortest glob path.
func NewShortestGlobPathCache() *ShortestGlobPathCache {
	return &ShortestGlobPathCache{
		Cache: make(map[string][]string),
	}
}

// FindShortestMatch finds the shortest match of the glob against the file.
func (s *ShortestGlobPathCache) FindShortestMatch(pattern string, file string) (string, error) {
	cached, globCachedBefore := s.Cache[pattern]
	if globCachedBefore {
		// Check if we have the shortest path cached already for the file.
		// E.g. the pattern /build/v* could have previously matched /build/v12/server.jar.
		// This would yield a cache entry of /build/v12 for the pattern /build/v*.
		// If a new file is tried by this method, e.g. /build/v12/client.jar, the cache is used in combination with a cheap strings.HasPrefix check
		// To determine the shortest match being /build/v12
		for _, cachedStart := range cached {
			if strings.HasPrefix(file, cachedStart) {
				return cachedStart, nil
			}
		}
	}

	compiledGlob, err := glob.Compile(pattern, '/')
	if err != nil {
		return "", fmt.Errorf("failed to compile pattern %s: %w", pattern, err)
	}

	directorySplit := strings.SplitAfter(file, "/")

	// Compute a slice that holds the / as individual entries.
	// /var/local/spellcore/ ends up as [/, var, /, local, /, spellcore, /]
	directorySplitIncludingSlashes := make([]string, 0, len(directorySplit)*2)
	for _, dir := range directorySplit {
		dirWithoutSlash := strings.TrimSuffix(dir, "/")
		if dirWithoutSlash != "" {
			directorySplitIncludingSlashes = append(directorySplitIncludingSlashes, dirWithoutSlash)
		}
		directorySplitIncludingSlashes = append(directorySplitIncludingSlashes, "/")
	}

	// Iterate over the directory split, the shortest one that satisfies the pattern is returned.
	var builder strings.Builder
	for _, pathPart := range directorySplitIncludingSlashes {
		builder.WriteString(pathPart)

		buildAsString := builder.String()
		matched := compiledGlob.Match(buildAsString)

		if !matched {
			continue
		}

		if !globCachedBefore {
			cached = make([]string, 0)
		}

		cached = append(cached, buildAsString)
		s.Cache[pattern] = cached

		return buildAsString, nil
	}

	return "", ErrFailedToFindShortestPath
}

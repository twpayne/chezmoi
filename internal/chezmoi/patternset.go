package chezmoi

import (
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

type patternSetIncludeType bool

const (
	patternSetInclude patternSetIncludeType = true
	patternSetExclude patternSetIncludeType = false
)

type patternSetMatchType int

const (
	patternSetMatchInclude patternSetMatchType = 1
	patternSetMatchUnknown patternSetMatchType = 0
	patternSetMatchExclude patternSetMatchType = -1
)

// An patternSet is a set of patterns.
type patternSet struct {
	includePatterns chezmoiset.Set[string]
	excludePatterns chezmoiset.Set[string]
}

// newPatternSet returns a new patternSet.
func newPatternSet() *patternSet {
	return &patternSet{
		includePatterns: chezmoiset.New[string](),
		excludePatterns: chezmoiset.New[string](),
	}
}

// LogValue implements log/slog.LogValuer.LogValue.
func (ps *patternSet) LogValue() slog.Value {
	if ps == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.Any("includePatterns", ps.includePatterns.Elements()),
		slog.Any("excludePatterns", ps.excludePatterns.Elements()),
	)
}

// add adds a pattern to ps.
func (ps *patternSet) add(pattern string, include patternSetIncludeType) error {
	if ok := doublestar.ValidatePattern(pattern); !ok {
		return fmt.Errorf("%s: invalid pattern", pattern)
	}
	switch include {
	case patternSetInclude:
		ps.includePatterns.Add(pattern)
	case patternSetExclude:
		ps.excludePatterns.Add(pattern)
	}
	return nil
}

// glob returns all matches in fileSystem.
func (ps *patternSet) glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	allMatches := chezmoiset.New[string]()
	for includePattern := range ps.includePatterns {
		matches, err := Glob(fileSystem, filepath.ToSlash(prefix+includePattern))
		if err != nil {
			return nil, err
		}
		allMatches.Add(matches...)
	}
	for match := range allMatches {
		for excludePattern := range ps.excludePatterns {
			exclude, err := doublestar.Match(path.Clean(prefix+excludePattern), match)
			if err != nil {
				return nil, err
			}
			if exclude {
				allMatches.Remove(match)
			}
		}
	}
	matchesSlice := allMatches.Elements()
	for i, match := range matchesSlice {
		matchesSlice[i] = filepath.ToSlash(match)[len(prefix):]
	}
	slices.Sort(matchesSlice)
	return matchesSlice, nil
}

// match returns if name matches ps.
func (ps *patternSet) match(name string) patternSetMatchType {
	// If name is explicitly excluded, then return exclude.
	for pattern := range ps.excludePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return patternSetMatchExclude
		}
	}

	// If name is explicitly included, then return include.
	for pattern := range ps.includePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return patternSetMatchInclude
		}
	}

	// If name did not match any include or exclude patterns...
	switch {
	case len(ps.includePatterns) > 0 && len(ps.excludePatterns) == 0:
		// ...only include patterns were specified, so exclude by default.
		return patternSetMatchExclude
	case len(ps.includePatterns) == 0 && len(ps.excludePatterns) > 0:
		// ...only exclude patterns were specified, so include by default.
		return patternSetMatchInclude
	default:
		// ...both include and exclude were specified, so return unknown.
		return patternSetMatchUnknown
	}
}

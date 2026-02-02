package chezmoi

import (
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/internal/chezmoiset"
)

type PatternSetIncludeType bool

const (
	PatternSetInclude PatternSetIncludeType = true
	PatternSetExclude PatternSetIncludeType = false
)

type PatternSetMatchType int

const (
	PatternSetMatchInclude PatternSetMatchType = 1
	PatternSetMatchUnknown PatternSetMatchType = 0
	PatternSetMatchExclude PatternSetMatchType = -1
)

// An PatternSet is a set of patterns.
type PatternSet struct {
	IncludePatterns chezmoiset.Set[string]
	ExcludePatterns chezmoiset.Set[string]
}

// NewPatternSet returns a new patternSet.
func NewPatternSet() *PatternSet {
	return &PatternSet{
		IncludePatterns: chezmoiset.New[string](),
		ExcludePatterns: chezmoiset.New[string](),
	}
}

// LogValue implements log/slog.LogValuer.LogValue.
func (ps *PatternSet) LogValue() slog.Value {
	if ps == nil {
		return slog.Value{}
	}
	return slog.GroupValue(
		slog.Any("includePatterns", slices.Sorted(ps.IncludePatterns.Elements())),
		slog.Any("excludePatterns", slices.Sorted(ps.ExcludePatterns.Elements())),
	)
}

// Add adds a pattern to ps.
func (ps *PatternSet) Add(pattern string, include PatternSetIncludeType) error {
	if ok := doublestar.ValidatePattern(pattern); !ok {
		return fmt.Errorf("%s: invalid pattern", pattern)
	}
	switch include {
	case PatternSetInclude:
		ps.IncludePatterns.Add(pattern)
	case PatternSetExclude:
		ps.ExcludePatterns.Add(pattern)
	}
	return nil
}

// Glob returns all matches in fileSystem.
func (ps *PatternSet) Glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	allMatches := chezmoiset.New[string]()
	for includePattern := range ps.IncludePatterns {
		matches, err := Glob(fileSystem, filepath.ToSlash(prefix+includePattern))
		if err != nil {
			return nil, err
		}
		allMatches.Add(matches...)
	}
	for match := range allMatches {
		for excludePattern := range ps.ExcludePatterns {
			exclude, err := doublestar.Match(path.Clean(prefix+excludePattern), match)
			if err != nil {
				return nil, err
			}
			if exclude {
				allMatches.Remove(match)
			}
		}
	}
	sortedMatches := slices.Sorted(allMatches.Elements())
	for i, match := range sortedMatches {
		sortedMatches[i] = filepath.ToSlash(match)[len(prefix):]
	}
	return sortedMatches, nil
}

// Match returns if name matches ps.
func (ps *PatternSet) Match(name string) PatternSetMatchType {
	// If name is explicitly excluded, then return exclude.
	for pattern := range ps.ExcludePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return PatternSetMatchExclude
		}
	}

	// If name is explicitly included, then return include.
	for pattern := range ps.IncludePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return PatternSetMatchInclude
		}
	}

	// If name did not match any include or exclude patterns...
	switch {
	case len(ps.IncludePatterns) > 0 && len(ps.ExcludePatterns) == 0:
		// ...only include patterns were specified, so exclude by default.
		return PatternSetMatchExclude
	case len(ps.IncludePatterns) == 0 && len(ps.ExcludePatterns) > 0:
		// ...only exclude patterns were specified, so include by default.
		return PatternSetMatchInclude
	default:
		// ...both include and exclude were specified, so return unknown.
		return PatternSetMatchUnknown
	}
}

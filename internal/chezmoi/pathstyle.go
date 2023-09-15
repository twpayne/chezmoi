package chezmoi

import (
	"fmt"
	"strings"
)

type PathStyle string

const (
	PathStyleAbsolute       PathStyle = "absolute"
	PathStyleRelative       PathStyle = "relative"
	PathStyleSourceAbsolute PathStyle = "source-absolute"
	PathStyleSourceRelative PathStyle = "source-relative"
)

var (
	PathStyleStrings = []string{
		PathStyleAbsolute.String(),
		PathStyleRelative.String(),
		PathStyleSourceAbsolute.String(),
		PathStyleSourceRelative.String(),
	}

	PathStyleFlagCompletionFunc = FlagCompletionFunc(PathStyleStrings)
)

// Set implements github.com/spf13/pflag.Value.Set.
func (p *PathStyle) Set(s string) error {
	uniqueAbbreviations := UniqueAbbreviations(PathStyleStrings)
	pathStyleStr, ok := uniqueAbbreviations[s]
	if !ok {
		return fmt.Errorf("%s: unknown path style", s)
	}
	*p = PathStyle(pathStyleStr)
	return nil
}

func (p PathStyle) String() string {
	return string(p)
}

// Type implements github.com/spf13/pflag.Value.Type.
func (p PathStyle) Type() string {
	return strings.Join(PathStyleStrings, "|")
}

func (p PathStyle) Copy() *PathStyle {
	return &p
}

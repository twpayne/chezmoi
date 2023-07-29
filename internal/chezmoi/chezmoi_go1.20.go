//go:build go1.20

package chezmoi

import "strings"

var (
	CutPrefix = strings.CutPrefix
	CutSuffix = strings.CutSuffix
)

//go:build go1.18
// +build go1.18

package chezmoi

import (
	"bytes"
	"strings"
)

// FIXME when Go 1.18 is the minimum supported Go version, replace these with
// {strings,bytes}.Cut.
var (
	CutBytes  = bytes.Cut
	CutString = strings.Cut
)

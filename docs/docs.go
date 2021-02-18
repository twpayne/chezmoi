// Package docs contains chezmoi's documentation.
package docs

import "embed"

// FS contains all docs.
//go:embed *.md
var FS embed.FS

// Package templates contains chezmoi's templates.
package templates

import "embed"

// FS contains all templates.
//
//go:embed *.sh
//go:embed *.tmpl
var FS embed.FS

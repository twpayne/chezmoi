// Package templates contains chezmoi's templates.
package templates

import _ "embed"

//go:embed COMMIT_MESSAGE.tmpl
var CommitMessageTmpl string

//go:embed install.sh
var InstallSH []byte

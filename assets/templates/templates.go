// Package templates contains chezmoi's templates.
package templates

import _ "embed"

//go:embed COMMIT_MESSAGE.tmpl
var CommitMessageTmpl string

//go:embed install.sh
var InstallSh []byte

//go:embed install-init-shell.sh.tmpl
var InstallInitShellShTmpl []byte

// Package templates contains chezmoi's templates.
package templates

import _ "embed"

//go:embed COMMIT_MESSAGE.tmpl
var CommitMessageTmpl []byte

//go:embed install.sh
var InstallSH []byte

//go:embed versioninfo.json.tmpl
var VersionInfoJSON []byte

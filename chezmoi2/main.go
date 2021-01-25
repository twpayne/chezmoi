//go:generate go run ../internal/cmd/generate-assets -o cmd/docs.gen.go -tags=!noembeddocs -trimprefix=../ ../docs/CHANGES.md ../docs/CONTRIBUTING.md ../docs/FAQ.md ../docs/HOWTO.md ../docs/INSTALL.md ../docs/MEDIA.md ../docs/QUICKSTART.md
//go:generate go run ../internal/cmd/generate-assets -o cmd/reference.gen.go -tags=!noembeddocs docs/REFERENCE.md
//go:generate go run ../internal/cmd/generate-assets -o cmd/templates.gen.go -trimprefix=../ ../assets/templates/COMMIT_MESSAGE.tmpl
//go:generate go run ../internal/cmd/generate-helps -o cmd/helps.gen.go -i docs/REFERENCE.md
//go:generate go run . completion bash -o completions/chezmoi2-completion.bash
//go:generate go run . completion fish -o completions/chezmoi2.fish
//go:generate go run . completion powershell -o completions/chezmoi2.ps1
//go:generate go run . completion zsh -o completions/chezmoi2.zsh

package main

import (
	"os"

	"github.com/twpayne/chezmoi/chezmoi2/cmd"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	if exitCode := cmd.Main(cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}, os.Args[1:]); exitCode != 0 {
		os.Exit(exitCode)
	}
}

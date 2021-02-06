//go:generate go run ./internal/cmd/generate-assets -o cmd/docs.gen.go -tags=!noembeddocs docs/CHANGES.md docs/COMPARISON.md docs/CONTRIBUTING.md docs/FAQ.md docs/HOWTO.md docs/INSTALL.md docs/MEDIA.md docs/QUICKSTART.md docs/REFERENCE.md docs/TEMPLATING.md
//go:generate go run ./internal/cmd/generate-assets -o cmd/templates.gen.go assets/templates/COMMIT_MESSAGE.tmpl
//go:generate go run ./internal/cmd/generate-helps -o cmd/helps.gen.go -i docs/REFERENCE.md
//go:generate go run . completion bash -o completions/chezmoi-completion.bash
//go:generate go run . completion fish -o completions/chezmoi.fish
//go:generate go run . completion powershell -o completions/chezmoi.ps1
//go:generate go run . completion zsh -o completions/chezmoi.zsh

package main

import (
	"fmt"
	"os"

	"github.com/twpayne/chezmoi/cmd"
)

var (
	version = ""
	commit  = ""
	date    = ""
	builtBy = ""
)

func run() error {
	cmd.VersionStr = version
	cmd.Commit = commit
	cmd.Date = date
	cmd.BuiltBy = builtBy
	return cmd.Execute()
}

func main() {
	if err := run(); err != nil {
		if s := err.Error(); s != "" {
			//nolint:forbidigo
			fmt.Printf("chezmoi: %s\n", s)
		}
		os.Exit(1)
	}
}

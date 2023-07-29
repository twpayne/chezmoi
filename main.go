//go:build go1.19

//go:generate go run . completion bash -o completions/chezmoi-completion.bash
//go:generate go run . completion fish -o completions/chezmoi.fish
//go:generate go run . completion powershell -o completions/chezmoi.ps1
//go:generate go run . completion zsh -o completions/chezmoi.zsh
//go:generate go run ./internal/cmds/generate-install.sh -o assets/scripts/install.sh
//go:generate go run ./internal/cmds/generate-install.sh -b .local/bin -o assets/scripts/install-local-bin.sh

package main

import (
	"os"

	"github.com/twpayne/chezmoi/v2/internal/cmd"
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

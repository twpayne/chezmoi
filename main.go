//go:generate go tool chezmoi completion bash -o completions/chezmoi-completion.bash
//go:generate go tool chezmoi completion fish -o completions/chezmoi.fish
//go:generate go tool chezmoi completion powershell -o completions/chezmoi.ps1
//go:generate go tool chezmoi completion zsh -o completions/chezmoi.zsh
//go:generate go tool generate-helps -o internal/cmd/helps.gen.go
//go:generate go tool generate-install.sh -o assets/scripts/install.sh
//go:generate go tool generate-install.sh -b .local/bin -o assets/scripts/install-local-bin.sh
//go:generate go tool generate-license -o internal/cmd/license.gen.go

// chezmoi manages your dotfiles across multiple machines, securely.
package main

import (
	"os"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/twpayne/chezmoi/internal/cmd"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

func main() {
	// Set GOMAXPROCS to match the Linux CPU quota.
	_, _ = maxprocs.Set()

	if exitCode := cmd.Main(cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}, os.Args[1:]); exitCode != 0 {
		os.Exit(exitCode)
	}
}

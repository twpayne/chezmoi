package main

import (
	"github.com/twpayne/chezmoi/cmd"
)

var (
	// These variables are set goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(cmd.Version{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
}

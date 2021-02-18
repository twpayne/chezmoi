package main

import (
	"os"

	"github.com/twpayne/chezmoi/cmd"
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

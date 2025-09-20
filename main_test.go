package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/cmd"
)

func TestMain(t *testing.T) {
	versionInfo := cmd.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
		BuiltBy: builtBy,
	}
	args := []string{"--version"}
	assert.Equal(t, 0, cmd.Main(versionInfo, args))
}

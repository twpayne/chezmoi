package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twpayne/chezmoi/v2/pkg/cmd"
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

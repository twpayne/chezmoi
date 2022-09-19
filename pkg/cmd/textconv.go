package cmd

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type textConvElement struct {
	Pattern string   `json:"pattern" toml:"pattern" yaml:"pattern"`
	Command string   `json:"command" toml:"command" yaml:"command"`
	Args    []string `json:"args" toml:"args" yaml:"args"`
}

type textConv []*textConvElement

func (t textConv) convert(absPath chezmoi.AbsPath, data []byte) ([]byte, error) {
	var longestPatternElement *textConvElement
	for _, command := range t {
		ok, err := doublestar.Match(command.Pattern, absPath.String())
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if longestPatternElement == nil || len(command.Pattern) > len(longestPatternElement.Pattern) {
			longestPatternElement = command
		}
	}
	if longestPatternElement == nil {
		return data, nil
	}

	cmd := exec.Command(longestPatternElement.Command, longestPatternElement.Args...) //nolint:gosec
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(cmd)
}

package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type textConvElement struct {
	Pattern string   `json:"pattern" mapstructure:"pattern" yaml:"pattern"`
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
}

type textConv []*textConvElement

func (t textConv) convert(path string, data []byte) ([]byte, error) {
	var longestPatternElement *textConvElement
	for _, command := range t {
		ok, err := doublestar.Match(command.Pattern, path)
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

	cmd := exec.Command(longestPatternElement.Command, longestPatternElement.Args...)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(slog.Default(), cmd)
}

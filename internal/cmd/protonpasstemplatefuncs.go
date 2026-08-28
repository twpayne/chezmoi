package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"

	"chezmoi.io/chezmoi/v2/internal/chezmoi"
	"chezmoi.io/chezmoi/v2/internal/chezmoilog"
)

type protonPassConfig struct {
	Command         string `json:"command" mapstructure:"command" yaml:"command"`
	outputCache     map[string][]byte
	attachmentCache map[string]string
}

func (c *Config) protonPassAttachmentTemplateFunc(shareID, itemID, attachmentID string) string {
	chezmoi.SkipTemplateIf(c.skipSecrets)

	key := shareID + "\x00" + itemID + "\x00" + attachmentID
	if contents, ok := c.ProtonPass.attachmentCache[key]; ok {
		return contents
	}

	tempDir := mustValue(c.tempDir("chezmoi-proton-pass"))
	outputFile := mustValue(os.CreateTemp(tempDir.String(), "attachment-*"))
	defer func() { _ = outputFile.Close() }()

	args := []string{
		"item", "attachment", "download",
		"--share-id", shareID,
		"--item-id", itemID,
		"--attachment-id", attachmentID,
		"--output", outputFile.Name(),
	}
	cmd := exec.Command(c.ProtonPass.Command, args...)
	// pass-cli is very chatty. By default, ignore its stdout and stderr, but
	// connect them if the --debug flag is passed.
	if c.debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	contents := string(mustValue(io.ReadAll(outputFile)))
	if c.ProtonPass.attachmentCache == nil {
		c.ProtonPass.attachmentCache = make(map[string]string)
	}
	c.ProtonPass.attachmentCache[key] = contents

	return contents
}

func (c *Config) protonPassTemplateFunc(item string) string {
	chezmoi.SkipTemplateIf(c.skipSecrets)

	args := []string{"item", "view", item}
	return string(mustValue(c.protonPassOutput(args)))
}

func (c *Config) protonPassJSONTemplateFunc(item string) any {
	chezmoi.SkipTemplateIf(c.skipSecrets)

	args := []string{"item", "view", item, "--output=json"}
	output := mustValue(c.protonPassOutput(args))

	var result map[string]any
	must(json.Unmarshal(output, &result))
	return result
}

func (c *Config) protonPassOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.ProtonPass.outputCache[key]; ok {
		return data, nil
	}

	cmd := exec.Command(c.ProtonPass.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.ProtonPass.outputCache == nil {
		c.ProtonPass.outputCache = make(map[string][]byte)
	}
	c.ProtonPass.outputCache[key] = output
	return output, nil
}

package cmd

import (
	"fmt"
	"os/exec"

	"github.com/coreos/go-semver/semver"
)

type cmdOutputError struct {
	path   string
	args   []string
	output []byte
	err    error
}

func newCmdOutputError(cmd *exec.Cmd, output []byte, err error) *cmdOutputError {
	return &cmdOutputError{
		path:   cmd.Path,
		args:   cmd.Args,
		output: output,
		err:    err,
	}
}

func (e *cmdOutputError) Error() string {
	if len(e.output) == 0 {
		return fmt.Sprintf("%s: %v", shellQuoteCommand(e.path, e.args[1:]), e.err)
	}
	return fmt.Sprintf("%s: %v\n%s", shellQuoteCommand(e.path, e.args[1:]), e.err, e.output)
}

func (e *cmdOutputError) Unwrap() error {
	return e.err
}

type extractVersionError struct {
	output []byte
}

func (e *extractVersionError) Error() string {
	return fmt.Sprintf("%s: cannot extract version", e.output)
}

type parseCmdOutputError struct {
	command string
	args    []string
	output  []byte
	err     error
}

func newParseCmdOutputError(
	command string,
	args []string,
	output []byte,
	err error,
) *parseCmdOutputError {
	return &parseCmdOutputError{
		command: command,
		args:    args,
		output:  output,
		err:     err,
	}
}

func (e *parseCmdOutputError) Error() string {
	return fmt.Sprintf("%s: %v\n%s", shellQuoteCommand(e.command, e.args), e.err, e.output)
}

func (e *parseCmdOutputError) Unwrap() error {
	return e.err
}

type parseVersionError struct {
	output []byte
	err    error
}

func (e *parseVersionError) Error() string {
	return fmt.Sprintf("%s: cannot parse version: %v", e.output, e.err)
}

func (e *parseVersionError) Unwrap() error {
	return e.err
}

type unsupportedVersionError struct {
	version *semver.Version
}

func (e *unsupportedVersionError) Error() string {
	return fmt.Sprintf("%s: unsupported version", e.version)
}

type versionTooOldError struct {
	have *semver.Version
	need *semver.Version
}

func (e *versionTooOldError) Error() string {
	return fmt.Sprintf("found version %s, need version %s or later", e.have, e.need)
}

package chezmoi

import (
	"os/exec"

	"github.com/rs/zerolog"
)

// An Interpreter interprets scripts.
type Interpreter struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// ExecCommand returns the *exec.Cmd to interpret name.
func (i *Interpreter) ExecCommand(name string) *exec.Cmd {
	if i.None() {
		return exec.Command(name)
	}
	return exec.Command(i.Command, append(i.Args, name)...) //nolint:gosec
}

// None returns if i represents no interpreter.
func (i *Interpreter) None() bool {
	return i == nil || i.Command == ""
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (i *Interpreter) MarshalZerologObject(event *zerolog.Event) {
	if i == nil {
		return
	}
	if i.Command != "" {
		event.Str("command", i.Command)
	}
	if i.Args != nil {
		event.Strs("args", i.Args)
	}
}

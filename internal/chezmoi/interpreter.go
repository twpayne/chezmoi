package chezmoi

import (
	"log/slog"
	"os/exec"
)

// An Interpreter interprets scripts.
type Interpreter struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
}

// ExecCommand returns the [*exec.Cmd] to interpret name.
func (i *Interpreter) ExecCommand(name string) *exec.Cmd {
	if i.None() {
		return exec.Command(name)
	}
	return exec.Command(i.Command, append(i.Args, name)...)
}

// None returns if i represents no interpreter.
func (i *Interpreter) None() bool {
	return i == nil || i.Command == ""
}

// LogValue implements log/slog.LogValuer.LogValue.
func (i *Interpreter) LogValue() slog.Value {
	var attrs []slog.Attr
	if i.Command != "" {
		attrs = append(attrs, slog.String("command", i.Command))
	}
	if i.Args != nil {
		attrs = append(attrs, slog.Any("args", i.Args))
	}
	return slog.GroupValue(attrs...)
}

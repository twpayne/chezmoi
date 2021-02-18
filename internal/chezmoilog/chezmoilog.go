// Package chezmoilog contains support for chezmoi logging.
package chezmoilog

import (
	"errors"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
)

// An OSExecCmdLogObject wraps an *os/exec.Cmd and adds
// github.com/rs/zerolog.LogObjectMarshaler functionality.
type OSExecCmdLogObject struct {
	*exec.Cmd
}

// An OSExecExitErrorLogObject wraps an error and adds
// github.com/rs/zerolog.LogObjectMarshaler functionality if the wrapped error
// is an os/exec.ExitError.
type OSExecExitErrorLogObject struct {
	Err error
}

// An OSProcessStateLogObject wraps an *os.ProcessState and adds
// github.com/rs/zerolog.LogObjectMarshaler functionality.
type OSProcessStateLogObject struct {
	*os.ProcessState
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (cmd OSExecCmdLogObject) MarshalZerologObject(event *zerolog.Event) {
	if cmd.Cmd == nil {
		return
	}
	if cmd.Path != "" {
		event.Str("path", cmd.Path)
	}
	if cmd.Args != nil {
		event.Strs("args", cmd.Args)
	}
	if cmd.Dir != "" {
		event.Str("dir", cmd.Dir)
	}
	if cmd.Env != nil {
		event.Strs("env", cmd.Env)
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (err OSExecExitErrorLogObject) MarshalZerologObject(event *zerolog.Event) {
	if err.Err == nil {
		return
	}
	var osExecExitError *exec.ExitError
	if !errors.As(err.Err, &osExecExitError) {
		return
	}
	event.EmbedObject(OSProcessStateLogObject{osExecExitError.ProcessState})
	if osExecExitError.Stderr != nil {
		event.Bytes("stderr", osExecExitError.Stderr)
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (p OSProcessStateLogObject) MarshalZerologObject(event *zerolog.Event) {
	if p.ProcessState == nil {
		return
	}
	if p.Exited() {
		if !p.Success() {
			event.Int("exitCode", p.ExitCode())
		}
	} else {
		event.Int("pid", p.Pid())
	}
	if userTime := p.UserTime(); userTime != 0 {
		event.Dur("userTime", userTime)
	}
	if systemTime := p.SystemTime(); systemTime != 0 {
		event.Dur("systemTime", systemTime)
	}
}

// FirstFewBytes returns the first few bytes of data in a human-readable form.
func FirstFewBytes(data []byte) []byte {
	const few = 64
	if len(data) > few {
		data = append([]byte{}, data[:few]...)
		data = append(data, '.', '.', '.')
	}
	return data
}

// LogCmdCombinedOutput calls cmd.CombinedOutput, logs the result, and returns the result.
func LogCmdCombinedOutput(logger zerolog.Logger, cmd *exec.Cmd) ([]byte, error) {
	combinedOutput, err := cmd.CombinedOutput()
	logger.Debug().
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		Err(err).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Bytes("combinedOutput", FirstFewBytes(combinedOutput)).
		Msg("CombinedOutput")
	return combinedOutput, err
}

// LogCmdOutput calls cmd.Output, logs the result, and returns the result.
func LogCmdOutput(logger zerolog.Logger, cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.Output()
	logger.Debug().
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		Err(err).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Bytes("output", FirstFewBytes(output)).
		Msg("Output")
	return output, err
}

// LogCmdRun calls cmd.Run, logs the result, and returns the result.
func LogCmdRun(logger zerolog.Logger, cmd *exec.Cmd) error {
	err := cmd.Run()
	logger.Debug().
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		Err(err).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Msg("Run")
	return err
}

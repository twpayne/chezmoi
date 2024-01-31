// Package chezmoilog contains support for chezmoi logging.
package chezmoilog

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const few = 64

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
	if osExecExitError := (&exec.ExitError{}); errors.As(err.Err, &osExecExitError) {
		event.EmbedObject(OSProcessStateLogObject{osExecExitError.ProcessState})
		if osExecExitError.Stderr != nil {
			event.Bytes("stderr", osExecExitError.Stderr)
			return
		}
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
	if len(data) > few {
		data = slices.Clone(data[:few])
		data = append(data, '.', '.', '.')
	}
	return data
}

// LogHTTPRequest calls httpClient.Do, logs the result to logger, and returns
// the result.
func LogHTTPRequest(logger *zerolog.Logger, client *http.Client, req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := client.Do(req)
	if resp != nil {
		logger.Err(err).
			Stringer("duration", time.Since(start)).
			Str("method", req.Method).
			Int64("size", resp.ContentLength).
			Int("statusCode", resp.StatusCode).
			Str("status", resp.Status).
			Stringer("url", req.URL).
			Msg("HTTPRequest")
	} else {
		logger.Err(err).
			Stringer("duration", time.Since(start)).
			Str("method", req.Method).
			Stringer("url", req.URL).
			Msg("HTTPRequest")
	}
	return resp, err
}

// LogCmdCombinedOutput calls cmd.CombinedOutput, logs the result, and returns the result.
func LogCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	combinedOutput, err := cmd.CombinedOutput()
	log.Err(err).
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Bytes("combinedOutput", Output(combinedOutput, err)).
		Stringer("duration", time.Since(start)).
		Int("size", len(combinedOutput)).
		Msg("CombinedOutput")
	return combinedOutput, err
}

// LogCmdOutput calls cmd.Output, logs the result, and returns the result.
func LogCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	output, err := cmd.Output()
	log.Err(err).
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Stringer("duration", time.Since(start)).
		Bytes("output", Output(output, err)).
		Int("size", len(output)).
		Msg("Output")
	return output, err
}

// LogCmdRun calls cmd.Run, logs the result, and returns the result.
func LogCmdRun(cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Run()
	log.Err(err).
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Stringer("duration", time.Since(start)).
		Msg("Run")
	return err
}

// LogCmdStart calls cmd.Start, logs the result, and returns the result.
func LogCmdStart(cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Start()
	log.Err(err).
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Time("start", start).
		Msg("Start")
	return err
}

// LogCmdWait calls cmd.Wait, logs the result, and returns the result.
func LogCmdWait(cmd *exec.Cmd) error {
	err := cmd.Wait()
	end := time.Now()
	log.Err(err).
		EmbedObject(OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(OSExecExitErrorLogObject{Err: err}).
		Time("end", end).
		Msg("Wait")
	return err
}

// Output returns the first few bytes of output if err is nil, otherwise it
// returns the full output.
func Output(data []byte, err error) []byte {
	if err != nil {
		return data
	}
	return FirstFewBytes(data)
}

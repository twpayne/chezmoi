// Package chezmoilog contains support for chezmoi logging.
package chezmoilog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"time"
)

const few = 64

// An OSExecCmdLogValuer wraps an [*os/exec.Cmd] and adds [log/slog.LogValuer]
// functionality.
type OSExecCmdLogValuer struct {
	*exec.Cmd
}

// An OSExecExitLogValuerError wraps an [*os/exec.ExitError] and adds
// [log/slog.LogValuer].
type OSExecExitLogValuerError struct {
	*exec.ExitError
}

// An OSProcessStateLogValuer wraps an [*os.ProcessState] and adds
// [log/slog.LogValuer] functionality.
type OSProcessStateLogValuer struct {
	*os.ProcessState
}

// LogValuer implements log/slog.LogValuer.LogValue.
func (cmd OSExecCmdLogValuer) LogValuer() slog.Value {
	var attrs []slog.Attr
	if cmd.Path != "" {
		attrs = append(attrs, slog.String("path", cmd.Path))
	}
	if len(cmd.Args) != 0 {
		attrs = append(attrs, slog.Any("args", cmd.Args))
	}
	if cmd.Dir != "" {
		attrs = append(attrs, slog.String("dir", cmd.Dir))
	}
	if len(cmd.Env) != 0 {
		attrs = append(attrs, slog.Any("env", cmd.Env))
	}
	return slog.GroupValue(attrs...)
}

// LogValuer implements log/slog.LogValuer.LogValue.
func (err OSExecExitLogValuerError) LogValuer() slog.Value {
	attrs := []slog.Attr{
		slog.Any("processState", OSProcessStateLogValuer{err.ProcessState}),
	}
	if osExecExitError := (&exec.ExitError{}); errors.As(err, &osExecExitError) {
		attrs = append(attrs, Bytes("stderr", err.Stderr))
	}
	return slog.GroupValue(attrs...)
}

// LogValue implements log/slog.LogValuer.LogValue.
func (p OSProcessStateLogValuer) LogValue() slog.Value {
	var attrs []slog.Attr
	if p.ProcessState != nil {
		if p.Exited() {
			if !p.Success() {
				attrs = append(attrs, slog.Int("exitCode", p.ExitCode()))
			}
		} else {
			attrs = append(attrs, slog.Int("pid", p.Pid()))
		}
		if userTime := p.UserTime(); userTime != 0 {
			attrs = append(attrs, slog.Duration("userTime", userTime))
		}
		if systemTime := p.SystemTime(); systemTime != 0 {
			attrs = append(attrs, slog.Duration("systemTime", systemTime))
		}
	}
	return slog.GroupValue(attrs...)
}

func AppendExitErrorAttrs(attrs []slog.Attr, err error) []slog.Attr {
	var execExitError *exec.ExitError
	if !errors.As(err, &execExitError) {
		return append(attrs, slog.Any("err", err))
	}

	if execExitError.ProcessState != nil {
		if execExitError.Exited() {
			attrs = append(attrs, slog.Int("exitCode", execExitError.ExitCode()))
		} else {
			attrs = append(attrs, slog.Int("pid", execExitError.Pid()))
		}
		if userTime := execExitError.UserTime(); userTime != 0 {
			attrs = append(attrs, slog.Duration("userTime", userTime))
		}
		if systemTime := execExitError.SystemTime(); systemTime != 0 {
			attrs = append(attrs, slog.Duration("systemTime", systemTime))
		}
	}

	return attrs
}

// Bytes returns an [slog.Attr] with the value data.
func Bytes(key string, data []byte) slog.Attr {
	return slog.String(key, string(data))
}

// FirstFewBytes returns an [slog.Attr] with the value of the first few bytes of
// data.
func FirstFewBytes(key string, data []byte) slog.Attr {
	return slog.String(key, string(firstFewBytesHelper(data)))
}

// LogHTTPRequest calls httpClient.Do, logs the result to logger, and returns
// the result.
func LogHTTPRequest(ctx context.Context, logger *slog.Logger, client *http.Client, req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := client.Do(req)
	attrs := []slog.Attr{
		slog.Duration("duration", time.Since(start)),
		slog.String("method", req.Method),
		Stringer("url", req.URL),
	}
	if resp != nil {
		attrs = append(attrs,
			slog.Int("statusCode", resp.StatusCode),
			slog.String("status", resp.Status),
			slog.Int("contentLength", int(resp.ContentLength)),
		)
	}
	InfoOrErrorContext(ctx, logger, "HTTPRequest", err, attrs...)
	return resp, err
}

// LogCmdCombinedOutput calls cmd.CombinedOutput, logs the result, and returns
// the result.
func LogCmdCombinedOutput(logger *slog.Logger, cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	combinedOutput, err := cmd.CombinedOutput()
	attrs := []slog.Attr{
		slog.Any("cmd", OSExecCmdLogValuer{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
		slog.Int("size", len(combinedOutput)),
		slog.Any("combinedOutput", firstFewBytesHelper(combinedOutput)),
	}
	attrs = AppendExitErrorAttrs(attrs, err)
	InfoOrErrorContext(context.Background(), logger, "Output", err, attrs...)
	return combinedOutput, err
}

// LogCmdOutput calls cmd.Output, logs the result, and returns the result.
func LogCmdOutput(logger *slog.Logger, cmd *exec.Cmd) ([]byte, error) {
	start := time.Now()
	output, err := cmd.Output()
	attrs := []slog.Attr{
		slog.Any("cmd", OSExecCmdLogValuer{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
		slog.Int("size", len(output)),
		slog.Any("output", firstFewBytesHelper(output)),
	}
	attrs = AppendExitErrorAttrs(attrs, err)
	InfoOrErrorContext(context.Background(), logger, "Output", err, attrs...)
	return output, err
}

// LogCmdRun calls cmd.Run, logs the result, and returns the result.
func LogCmdRun(logger *slog.Logger, cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Run()
	attrs := []slog.Attr{
		slog.Any("cmd", OSExecCmdLogValuer{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
	}
	attrs = AppendExitErrorAttrs(attrs, err)
	InfoOrErrorContext(context.Background(), logger, "Run", err, attrs...)
	return err
}

// LogCmdStart calls cmd.Start, logs the result, and returns the result.
func LogCmdStart(logger *slog.Logger, cmd *exec.Cmd) error {
	start := time.Now()
	err := cmd.Start()
	attrs := []slog.Attr{
		slog.Any("cmd", OSExecCmdLogValuer{Cmd: cmd}),
		slog.Time("start", start),
	}
	attrs = AppendExitErrorAttrs(attrs, err)
	InfoOrErrorContext(context.Background(), logger, "Start", err, attrs...)
	return err
}

// LogCmdWait calls cmd.Wait, logs the result, and returns the result.
func LogCmdWait(logger *slog.Logger, cmd *exec.Cmd) error {
	err := cmd.Wait()
	end := time.Now()
	attrs := []slog.Attr{
		slog.Any("cmd", OSExecCmdLogValuer{Cmd: cmd}),
		slog.Time("end", end),
	}
	attrs = AppendExitErrorAttrs(attrs, err)
	InfoOrError(logger, "Wait", err, attrs...)
	return err
}

func InfoOrError(logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	InfoOrErrorContext(context.Background(), logger, msg, err, attrs...)
}

func InfoOrErrorContext(ctx context.Context, logger *slog.Logger, msg string, err error, attrs ...slog.Attr) {
	if logger == nil {
		return
	}
	args := make([]any, 0, len(attrs)+1)
	if err != nil {
		args = append(args, slog.Any("err", err))
	}
	for _, attr := range attrs {
		args = append(args, attr)
	}
	level := slog.LevelInfo
	if err != nil {
		level = slog.LevelError
	}
	logger.Log(ctx, level, msg, args...)
}

// Stringer returns an [slog.Attr] with value.
func Stringer(key string, value fmt.Stringer) slog.Attr {
	return slog.String(key, value.String())
}

// firstFewBytesHelper returns the first few bytes of data.
func firstFewBytesHelper(data []byte) []byte {
	if len(data) > few {
		data = slices.Clone(data[:few])
		data = append(data, '.', '.', '.')
	}
	return data
}

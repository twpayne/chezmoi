package chezmoilog

import (
	"context"
	"log/slog"
)

// A NullHandler implements log/slog.Handler and drops all output.
type NullHandler struct{}

func (NullHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (NullHandler) Handle(context.Context, slog.Record) error { return nil }
func (h NullHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h NullHandler) WithGroup(string) slog.Handler           { return h }

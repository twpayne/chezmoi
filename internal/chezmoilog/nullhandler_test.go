package chezmoilog

import "log/slog"

var _ slog.Handler = NullHandler{}

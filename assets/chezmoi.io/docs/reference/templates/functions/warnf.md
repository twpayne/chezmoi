# `warnf` *format* [*arg*...]

`warnf` prints a message to stderr prefixed by `chezmoi: warning: ` and returns
the empty string. *format* is interpreted as a [printf-style format string][fmt]
with the given *arg*s.

[fmt]: https://pkg.go.dev/fmt#hdr-Printing

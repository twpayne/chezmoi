# `debugf` *format* [*arg*...]

`debugf` prints a message to stderr prefixed by `chezmoi: debug:` if the
`--verbose` flag is set. If `--verbose` is not set it does nothing. It returns
the empty string. *format* is interpreted as a [printf-style format string][fmt]
with the given *arg*s.

[fmt]: https://pkg.go.dev/fmt#hdr-Printing

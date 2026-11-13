# Tools

chezmoi includes a number of development and profiling tools controlled by the
`CHEZMOIDEV` environment variable.

The value of the `CHEZMOIDEV` environment variable is a comma-separated list of
`key=value` pairs. The following pairs are supported:

## `cpuprofile=`*filename*

Write a [Go CPU profile][pprof] to *filename*. Afterwards, *filename* can be
analyzed with [`go tool pprof`][pprof-tool].

## `ignoreflags=1`

By default, chezmoi includes a runtime check that each command's flags are
documented in the `## Flags` section of
`assets/chezmoi.io/docs/reference/commands/$COMMAND.md`. Setting
`CHEZMOIDEV=ignoreflags=1` disables this check, which can be useful when adding
new flags.

## `ignorehelp=1`

By default, chezmoi includes a runtime check that all commands have a long help
message extracted from the top level description of the command in
`assets/chezmoi.io/docs/reference/commands/$COMMAND.md`. Setting
`CHEZMOIDEV=ignorehelp=1` disables this check, which can be useful when adding
new commands.

## `memprofile=`*filename*

Write a [Go memory profile][pprof] to *filename*. Afterwards, *filename* can be
analyzed with [`go tool pprof`][pprof-tool].

## `trace=`*filename*

Write a [Go trace][trace] to *filename*. Afterwards, *filename* can be analyzed
with [`go tool trace`][trace-tool].

## Example

```console
$ CHEZMOIDEV=ignoreflags=1,trace=trace.out go tool chezmoi --version
...
$ go tool trace trace.out
```

[pprof]: https://pkg.go.dev/runtime/pprof
[pprof-tool]: https://github.com/google/pprof/blob/main/doc/README.md
[trace]: https://pkg.go.dev/runtime/trace
[trace-tool]: https://pkg.go.dev/cmd/trace

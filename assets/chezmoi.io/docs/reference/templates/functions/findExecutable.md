# `findExecutable` *file* *path-list*

`findExecutable` searches for an executable named *file* in directories
identified by *path-list*. The result will be the executable file concatenated
with the matching path. If an executable *file* cannot be found in *path-list*,
`findExecutable` returns an empty string.

`findExecutable` is provided as an alternative to
[`lookPath`](/reference/templates/functions/lookPath) so that you can
interrogate the system PATH as it would be configured after `chezmoi apply`.
Like `lookPath`, `findExecutable` is not hermetic: its return value depends on
the state of the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

The return value of the first successful call to `findExecutable` is cached, and
future calls to `findExecutable` with the same parameters will return this path.

!!! info

    On Windows, the resulting path will contain the first found executable
    extension as identified by the environment variable `%PathExt%`.

!!! example

    ```
    {{ if findExecutable "rtx" (list "bin" "go/bin" ".cargo/bin" ".local/bin") }}
    # $HOME/.cargo/bin/rtx exists and will probably be in $PATH after apply
    {{ end }}
    ```

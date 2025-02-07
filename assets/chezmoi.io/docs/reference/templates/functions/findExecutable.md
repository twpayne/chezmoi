# `findExecutable` _file_ _path-list_

`findExecutable` searches for an executable named _file_ in directories
identified by _path-list_. The result will be the executable file concatenated
with the matching path. If an executable _file_ cannot be found in _path-list_,
`findExecutable` returns an empty string.

`findExecutable` is provided as an alternative to [`lookPath`][site-lookpath] so
that you can interrogate the system PATH as it would be configured after
`chezmoi apply`. Like `lookPath`, `findExecutable` is not hermetic: its return
value depends on the state of the file system at the moment the template is
executed. Exercise caution when using it in your templates.

The return value of the first successful call to `findExecutable` is cached, and
future calls to `findExecutable` with the same parameters will return this path.

!!! info

    On Windows, the resulting path will contain the first found executable
    extension as identified by the environment variable `%PathExt%`.

!!! example

    ```
    {{ if findExecutable "mise" (list "bin" "go/bin" ".cargo/bin" ".local/bin") }}
    # $HOME/.cargo/bin/mise exists and will probably be in $PATH after apply
    {{ end }}
    ```

[site-lookpath]: /reference/templates/functions/lookPath.md

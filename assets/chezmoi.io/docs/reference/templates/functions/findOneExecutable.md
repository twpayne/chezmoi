# `findOneExecutable` *file-list* *path-list*

`findOneExecutable` searches for an executable from *file-list* in directories
identified by *path-list*, finding the first matching executable in the first
matching directory (each directory is searched for matching executables in
turn). The result will be the executable file concatenated with the matching
path. If an executable from *file-list* cannot be found in *path-list*,
`findOneExecutable` returns an empty string.

`findOneExecutable` is provided as an alternative to
[`lookPath`](/reference/templates/functions/lookPath) so that you can
interrogate the system PATH as it would be configured after `chezmoi apply`.
Like `lookPath`, `findOneExecutable` is not hermetic: its return value depends
on the state of the filesystem at the moment the template is executed. Exercise
caution when using it in your templates.

The return value of the first successful call to `findOneExecutable` is cached,
and future calls to `findOneExecutable` with the same parameters will return
this path.

!!! info

    On Windows, the resulting path will contain the first found executable
    extension as identified by the environment variable `%PathExt%`.

!!! example

    ```
    {{ if findOneExecutable (list "eza" "exa") (list "bin" "go/bin" ".cargo/bin" ".local/bin") }}
    # $HOME/.cargo/bin/exa exists and will probably be in $PATH after apply
    {{ end }}
    ```

# `stat` *name*

`stat` runs [`os.Stat`](https://pkg.go.dev/os#File.Stat) on *name*. If *name*
exists it returns structured data. If *name* does not exist then it returns a
false value. If `os.Stat` returns any other error then it raises an error. The
structured value returned if *name* exists contains the fields `name`, `size`,
`mode`, `perm`, `modTime`, and `isDir`.

`stat` is not hermetic: its return value depends on the state of the filesystem
at the moment the template is executed. Exercise caution when using it in your
templates.

!!! example

    ```
    {{ if stat (joinPath .chezmoi.homeDir ".pyenv") }}
    # ~/.pyenv exists
    {{ end }}
    ```

# `lstat` *name*

`lstat` runs [`os.Lstat`][lstat] on *name*. If *name* exists it returns
structured data. If *name* does not exist then it returns a false value. If
`os.Lstat` returns any other error then it raises an error. The structured value
returned if *name* exists contains the fields `name`, `size`, `mode`, `perm`,
`modTime`, `isDir`, and `type`.

`lstat` is not hermetic: its return value depends on the state of the file
system at the moment the template is executed. Exercise caution when using it in
your templates.

!!! example

    ```
    {{ if eq (joinPath .chezmoi.homeDir ".xinitrc" | lstat).type "symlink" }}
    # ~/.xinitrc exists and is a symlink
    {{ end }}
    ```

[lstat]: https://pkg.go.dev/os#File.Lstat

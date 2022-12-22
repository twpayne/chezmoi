# `managed` [*path*...]

List all managed entries in the destination directory under all *path*s in
alphabetical order. When no *path*s are supplied, list all managed entries in
the destination directory in alphabetical order.

## `-p`, `--path-style` `absolute`|`relative`|`source-absolute`|`source-relative`

Print paths in the given style. Relative paths are relative to the destination
directory. The default is `relative`.

!!! example

    ```console
    $ chezmoi managed
    $ chezmoi managed --include=files
    $ chezmoi managed --include=files,symlinks
    $ chezmoi managed -i dirs
    $ chezmoi managed -i dirs,files
    $ chezmoi managed -i files ~/.config
    $ chezmoi managed --exclude=encrypted --path-style=source-relative
    ```

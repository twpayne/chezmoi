# `managed` [*path*...]

List all managed entries in the destination directory under all *path*s in
alphabetical order. When no *path*s are supplied, list all managed entries in
the destination directory in alphabetical order.

!!! example

    ```console
    $ chezmoi managed
    $ chezmoi managed --include=files
    $ chezmoi managed --include=files,symlinks
    $ chezmoi managed -i dirs
    $ chezmoi managed -i dirs,files
    $ chezmoi managed -i files ~/.config
    $ chezmoi managed --exclude=encrypted
    ```

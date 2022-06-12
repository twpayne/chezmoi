# `managed` [*target*...]

List all managed entries in the destination directory under all *target*s in
alphabetical order.  When no *target*s are supplied, list all managed entries in
the destination directory in alphabetical order.

## `-i`, `--include` *types*

Only include entries of type *types*.

!!! example

    ```console
    $ chezmoi managed
    $ chezmoi managed --include=files
    $ chezmoi managed --include=files,symlinks
    $ chezmoi managed -i dirs
    $ chezmoi managed -i dirs,files
    $ chezmoi managed -i files ~/.config
    ```

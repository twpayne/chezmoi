# `unmanaged` [*path*...]

List all unmanaged files in *path*s. When no *path*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *path*s that are not found on the filesystem.

## `-p`, `--path-style` `absolute`|`relative`

Print paths in the given style. Relative paths are relative to the destination
directory. The default is `relative`.

!!! example

    ```console
    $ chezmoi unmanaged
    $ chezmoi unmanaged ~/.config/chezmoi ~/.ssh
    ```

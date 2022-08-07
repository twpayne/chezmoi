# `unmanaged` [*path*...]

List all unmanaged files in *path*s. When no *path*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *path*s that are not found on the filesystem.

!!! example

    ```console
    $ chezmoi unmanaged
    $ chezmoi unmanaged ~/.config/chezmoi ~/.ssh
    ```

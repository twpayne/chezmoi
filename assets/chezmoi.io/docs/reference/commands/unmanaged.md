# `unmanaged` [*target*...]

List all unmanaged files in *target*s.  When no *target*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *target*s that are not found on the filesystem.

!!! example

    ```console
    $ chezmoi unmanaged
    $ chezmoi unmanaged ~/.config/chezmoi ~/.ssh
    ```

# `cat` *target*...

Write the target contents of *target*s to stdout. *target*s must be files,
scripts, or symlinks. For files, the target file contents are written. For
scripts, the script's contents are written. For symlinks, the target is
written.

!!! example

    ```console
    $ chezmoi cat ~/.bashrc
    ```

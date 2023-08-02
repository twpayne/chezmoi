# `re-add` [*target*...]

Re-add modified files in the target state, preserving any `encrypted_`
attributes. chezmoi will not overwrite templates, and all entries that are not
files are ignored.

If no *target*s are specified then all modified files are re-added. If one or
more *target*s are given then only those targets are re-added.

!!! hint

    If you want to re-add a single file unconditionally, use `chezmoi add --force` instead.

!!! example

    ```console
    $ chezmoi re-add
    $ chezmoi re-add ~/.bashrc
    ```

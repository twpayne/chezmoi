# `verify` [*target*...]

Verify that all *target*s match their target state. chezmoi exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.

## `-i`, `--include` *types*

Only include entries of type *types*.

!!! example

    ```console
    $ chezmoi verify
    $ chezmoi verify ~/.bashrc
    ```

# `dump` [*target*...]

Dump the target state of *target*s. If no targets are specified, then the
entire target state.

## `-f`, `--format` `json`|`yaml`

Set the output format.

## `-i`, `--include` *types*

Only include entries of type *types*.

!!! example

    ```console
    $ chezmoi dump ~/.bashrc
    $ chezmoi dump --format=yaml
    ```

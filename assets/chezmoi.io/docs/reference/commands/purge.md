# `purge`

Remove chezmoi's configuration, state, and source directory, but leave the
target state intact.

## `-P`, `--binary`

Purge chezmoi binary.

## `-f`, `--force`

Remove without prompting.

!!! example

    ```console
    $ chezmoi purge
    $ chezmoi purge --force
    ```

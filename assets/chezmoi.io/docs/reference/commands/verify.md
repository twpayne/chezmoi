# `verify` [*target*...]

Verify that all *target*s match their target state. chezmoi exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),  defaults to `none`.

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

!!! example

    ```console
    $ chezmoi verify
    $ chezmoi verify ~/.bashrc
    ```

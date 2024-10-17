# `dump` [*target*...]

Dump the target state of *target*s. If no targets are specified, then the
entire target state.

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),  defaults to `none`.

## `-f`, `--format` `json`|`yaml`

Set the output format, default to `json`.

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

!!! example

    ```console
    $ chezmoi dump ~/.bashrc
    $ chezmoi dump --format=yaml
    ```

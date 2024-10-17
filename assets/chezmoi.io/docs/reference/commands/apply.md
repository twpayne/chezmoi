# `apply` [*target*...]

Ensure that *target*... are in the target state, updating them if necessary. If
no targets are specified, the state of all targets are ensured. If a target has
been modified since chezmoi last wrote it then the user will be prompted if
they want to overwrite the file.

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),  defaults to `none`.

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

## `--source-path`

Specify targets by source path, rather than target path. This is useful for
applying changes after editing.

!!! example

    ```console
    $ chezmoi apply
    $ chezmoi apply --dry-run --verbose
    $ chezmoi apply ~/.bashrc
    ```

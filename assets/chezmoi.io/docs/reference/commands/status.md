# `status`

Print the status of the files and scripts managed by chezmoi in a format
similar to [`git status`](https://git-scm.com/docs/git-status).

The first column of output indicates the difference between the last state
written by chezmoi and the actual state. The second column indicates the
difference between the actual state and the target state, and what effect
running [`chezmoi apply`](apply.md) will have.

| Character | Meaning   | First column       | Second column          |
| --------- | --------- | ------------------ | ---------------------- |
| Space     | No change | No change          | No change              |
| `A`       | Added     | Entry was created  | Entry will be created  |
| `D`       | Deleted   | Entry was deleted  | Entry will be deleted  |
| `M`       | Modified  | Entry was modified | Entry will be modified |
| `R`       | Run       | Not applicable     | Script will be run     |

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),  defaults to `none`.

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `-p`, `--path-style` `absolute`|`relative`|`source-absolute`|`source-relative`

Print paths in the given style. Relative paths are relative to the destination
directory. The default is `relative`.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

!!! example

    ```console
    $ chezmoi status
    ```

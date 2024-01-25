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

## `-i`, `--include` *types*

Only include entries of type *types*.

!!! example

    ```console
    $ chezmoi status
    ```

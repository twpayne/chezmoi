# `status`

Print the status of the files and scripts managed by chezmoi in a format
similar to [`git status`](https://git-scm.com/docs/git-status).

The first column of output indicates the difference between the last state
written by chezmoi and the actual state. The second column indicates the
difference between the actual state and the target state.

## `-i`, `--include` *types*

Only include entries of type *types*.

!!! example

    ```console
    $ chezmoi status
    ```

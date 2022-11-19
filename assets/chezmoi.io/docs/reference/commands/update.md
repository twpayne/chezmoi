# `update`

Pull changes from the source repo and apply any changes.

If `update.command` is set then chezmoi will run `update.command` with
`update.args` in the working tree. Otherwise, chezmoi will run `git pull
--autostash --rebase [--recurse-submodules]` , using chezmoi's builtin git if
`useBuiltinGit` is `true` or if `git.command` cannot be found in `$PATH`.

## `-i`, `--include` *types*

Only update entries of type *types*.

## `--recurse-submodules` *bool*

Update submodules recursively. This defaults to `true`.

!!! example

    ```console
    $ chezmoi update
    ```

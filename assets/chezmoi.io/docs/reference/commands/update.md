# `update`

Pull changes from the source repo and apply any changes.

If `update.command` is set then chezmoi will run `update.command` with
`update.args` in the working tree. Otherwise, chezmoi will run `git pull
--autostash --rebase [--recurse-submodules]` , using chezmoi's builtin git if
`useBuiltinGit` is `true` or if `git.command` cannot be found in `$PATH`.

## `--apply`

Apply changes after pulling, `true` by default. Can be disabled with `--apply=false`.

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#available-types),  defaults to `none`.

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `--recurse-submodules`

Update submodules recursively. This defaults to `true`. Can be disabled with `--recurse-submodules=false`.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

!!! example

    ```console
    $ chezmoi update
    ```

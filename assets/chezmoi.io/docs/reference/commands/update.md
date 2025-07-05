# `update`

Pull changes from the source repo and apply any changes.

If `update.command` is set then chezmoi will run `update.command` with
`update.args` in the working tree. Otherwise, chezmoi will run `git pull
--autostash --rebase [--recurse-submodules]` , using chezmoi's builtin git if
`useBuiltinGit` is `true` or if `git.command` cannot be found in `$PATH`.

## Flags

### `-a`, `--apply`

Apply changes after pulling, `true` by default. Can be disabled with `--apply=false`.

### `--recurse-submodules`

Update submodules recursively. This defaults to `true`. Can be disabled with `--recurse-submodules=false`.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

--8<-- "common-flags/init.md"

### `-P`, `--parent-dirs`

--8<-- "common-flags/parent-dirs.md"

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-true"

### `--backup`

--8<-- "common-flags/backup.md"


## Examples

```sh
chezmoi update
```

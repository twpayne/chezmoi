# `verify` [*target*...]

Verify that all *target*s match their target state. chezmoi exits with code 0
(success) if all targets match their target state, or 1 (failure) otherwise. If
no targets are specified then all targets are checked.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

Recreate config file from template.

### `-P`, `--parent-dirs`

Also perform command on all parent directories of *target*.

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

## Examples

```console
$ chezmoi verify
$ chezmoi verify ~/.bashrc
```

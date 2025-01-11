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

--8<-- "common-flags/init.md"

### `-P`, `--parent-dirs`

--8<-- "common-flags/parent-dirs.md"

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-true"

## Examples

```sh
chezmoi verify
chezmoi verify ~/.bashrc
```

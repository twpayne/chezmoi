# `dump` [*target*...]

Dump the target state of *target*s. If no targets are specified, then the
entire target state.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--format` `json`|`yaml`

--8<-- "common-flags/format.md"

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
chezmoi dump ~/.bashrc
chezmoi dump --format=yaml
```

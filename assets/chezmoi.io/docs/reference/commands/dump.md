# `dump` [*target*...]

Dump the target state of *target*s. If no targets are specified, then the
entire target state.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--format` `json`|`yaml`

Set the output format, default to `json`.

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
$ chezmoi dump ~/.bashrc
$ chezmoi dump --format=yaml
```

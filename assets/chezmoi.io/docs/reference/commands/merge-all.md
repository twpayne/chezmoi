# `merge-all`

Perform a three-way merge for file whose actual state does not match its target
state. The merge is performed with `chezmoi merge`.

## Common flags

### `--init`

Recreate config file from template.

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

## Examples

```console
$ chezmoi merge-all
```

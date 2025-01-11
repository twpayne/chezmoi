# `merge-all`

Perform a three-way merge for file whose actual state does not match its target
state. The merge is performed with `chezmoi merge`.

## Common flags

### `--init`

--8<-- "common-flags/init.md"

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-true"

## Examples

```sh
chezmoi merge-all
```

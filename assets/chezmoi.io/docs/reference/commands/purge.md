# `purge`

Remove chezmoi's configuration, state, and source directory, but leave the
target state intact.

## Flags

### `-P`, `--binary`

Purge chezmoi binary.

## Common flags

### `--force`

Remove without prompting.

## Examples

```console
$ chezmoi purge
$ chezmoi purge --force
```

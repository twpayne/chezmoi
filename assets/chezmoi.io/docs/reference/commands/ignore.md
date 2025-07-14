# `ignore`

Add, remove, activate, or deactivate patterns in `.chezmoiignore`.

## Subcommands

### `add [pattern]...`

Add patterns to `.chezmoiignore`. If the pattern already matches an ignored
entry, the command fails unless `-f`/`--force` is specified. If the
only argument is `-` then patterns are read from stdin until EOF, ignoring empty
lines.

### `remove [pattern]...`

Remove patterns from `.chezmoiignore`. If removing a pattern would still leave
the entry ignored due to another pattern, the command fails unless
`-f`/`--force` is specified. If the only argument is `-` then
patterns are read from stdin until EOF, ignoring empty lines.

### `activate [pattern]...`

Uncomment patterns in `.chezmoiignore`. If a pattern does not already exist the
command fails.

### `deactivate [pattern]...`

Comment patterns out in `.chezmoiignore`. If a pattern does not already exist.

### `query [name]...`

Print active patterns that match *name*. If none match then the command exits
with a non-zero status. If the only argument is `-` then names are read
from stdin until EOF, ignoring empty lines.

## Flags

### `-f`, `--force`

Force the operation even if the entry is already in the desired state.

## Examples

```sh
chezmoi ignore add '*.log'
chezmoi ignore remove README.md
chezmoi ignore activate '*.tmp'
chezmoi ignore deactivate notes.md
chezmoi ignore query foo.log
```

# `ssh` *host* *init-args*...

!!! Warning

    `ssh` is an experimental, potentially destructive, command.

SSH to *host*, install chezmoi, run `chezmoi init --apply *init-args*`, and
executes your shell.

## Flags

### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

## Examples

```sh
chezmoi ssh $HOSTNAME $GITHUB_USERNAME
chezmoi ssh $HOSTNAME -- --one-shot $GITHUB_USERNAME
```

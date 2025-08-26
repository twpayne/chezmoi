# `ssh` *host* *init-args*...

!!! Warning

    `ssh` is an experimental, potentially destructive, command.

SSH to *host*, install chezmoi, run `chezmoi init --apply *init-args*`, and
executes your shell.

## Flags

### `-p`, `--package`

Install chezmoi using the distribution's package manager, if possible.
Otherwise, fall back to `curl` or `wget` installation. If neither `curl` nor
`wget` are installed then install them with the distribution's package manager.

### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

## Examples

```sh
chezmoi ssh $HOSTNAME $GITHUB_USERNAME
chezmoi ssh $HOSTNAME -- --one-shot $GITHUB_USERNAME
```

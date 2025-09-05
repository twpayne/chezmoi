# `ssh` *host* *init-args*...

!!! Warning

    `ssh` is an experimental, potentially destructive, command.

SSH to *host*, install chezmoi, run `chezmoi init --apply *init-args*`, and
executes your shell.

## Flags

### `-p`, `--package-manager` *package-manager*

Install chezmoi using *package-manager*, if possible. Valid values for
*package-manager* include `apk`, `apt-get`, `brew`, `dnf`, `nix-env`, `pacman`,
`port`, `pkg`, `rpm`, `snap`, `xbps-install`, and `zypper`. Otherwise, fall back
to `curl` or `wget` installation. If neither `curl` nor `wget` are installed
then install them with *package-manager*.

### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

## Examples

```sh
chezmoi ssh $HOSTNAME $GITHUB_USERNAME
chezmoi ssh $HOSTNAME -- --one-shot $GITHUB_USERNAME
```

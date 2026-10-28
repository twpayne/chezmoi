# `docker`

<!-- markdownlint-disable no-duplicate-heading -->

!!! Hint

    To use [podman](https://podman.io/) instead of Docker, set the
    `docker.command` configuration variable to `podman`. You can also
    use the [`chezmoi podman` command][podman-command] instead of
    `chezmoi docker`.

Install chezmoi, run [`chezmoi init --apply`][init-command], and optionally
execute your shell in [Docker](https://docker.com/) containers.

## Subcommands

### `exec` *container-id* *init-args*...

Install chezmoi, run [`chezmoi init --apply *init-args*`][init-command], and
execute your shell in the existing Docker container *container-id*.

#### Flags

##### `-i`, `--interactive`

Keep stdin open even if not attached.

##### `-p`, `--package-manager` *package-manager*

Install chezmoi using *package-manager*, if possible. Valid values for
*package-manager* include `apk`, `apt-get`, `brew`, `dnf`, `nix-env`, `pacman`,
`port`, `pkg`, `rpm`, `snap`, `xbps-install`, and `zypper`. Otherwise, fall back
to `curl` or `wget` installation. If neither `curl` nor `wget` are installed
then install them with *package-manager*.

##### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

### `run` *image-id* *init-args*...

Create a new Docker container using *image-id*, and in it, install chezmoi, run
[`chezmoi init --apply *init-args*`][init-command], and execute your shell.

#### Flags

##### `-p`, `--package-manager` *package-manager*

Install chezmoi using *package-manager*, if possible. Valid values for
*package-manager* include `apk`, `apt-get`, `brew`, `dnf`, `nix-env`, `pacman`,
`port`, `pkg`, `rpm`, `snap`, `xbps-install`, and `zypper`. Otherwise, fall back
to `curl` or `wget` installation. If neither `curl` nor `wget` are installed
then install them with *package-manager*.

## Examples

```sh
chezmoi docker exec $CONTAINER_ID $YOUR_GITHUB_USERNAME
chezmoi docker run -p apk alpine:latest $YOUR_GITHUB_USERNAME
```

[init-command]: /reference/commands/init.md
[podman-command]: /reference/commands/podman.md

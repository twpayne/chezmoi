# `podman`

!!! Warning

    `podman` is an experimental command.

Install chezmoi, run `chezmoi init --apply`, and optionally execute your shell
in podman containers.

## Subcommands

### `exec` *container-id* *init-args*...

Install chezmoi, run `chezmoi init --apply *init-args*`, and execute your shell
in the existing podman container *container-id*.

#### Flags

#### `-i`, `--interactive`

Keep stdin open even if not attached.

#### `-p`, `--package`

Install chezmoi using the distribution's package manager, if possible.
Otherwise, fall back to `curl` or `wget` installation. If neither `curl` nor
`wget` are installed then install them with the distribution's package manager.

#### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

### `run` *image-id* *init-args*...

Create a new podman container using *image-id*, and in it, install chezmoit, run
`chezmoi init --apply *init-args*`, and execute your shell.

#### `-p`, `--package`

Install chezmoi using the distribution's package manager, if possible.
Otherwise, fall back to `curl` or `wget` installation. If neither `curl` nor
`wget` are installed then install them with the distribution's package manager.

## Examples

```sh
chezmoi podman exec $CONTAINER_ID $GITHUB_USERNAME
chezmoi podman run alpine:latest $GITHUB_USERNAME
```

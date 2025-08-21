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

FIXME describe flag

#### `-s`, `--shell`

After installing chezmoi, initializing your dotfiles, execute your shell. This
is the default.

### `run` *image-id* *init-args*...

Create a new podman container using *image-id*, and in it, install chezmoit, run
`chezmoi init --apply *init-args*`, and execute your shell.

## Examples

```sh
chezmoi podman run alpine:latest $GITHUB_USERNAME
chezmoi podman exec $CONTAINER_ID $GITHUB_USERNAME
```

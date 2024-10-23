# `archive` [*target*....]

Generate an archive of the target state, or only the targets specified. This
can be piped into `tar` to inspect the target state.

## Flags

### `-f`, `--format` *format*

Write the archive in *format*. If `--output` is set the format is guessed from
the extension, otherwise the default is `tar`.

| Supported formats |
| ----------------- |
| `tar`             |
| `tar.bz2`         |
| `tar.gz`          |
| `tar.xz`          |
| `tar.zst`         |
| `tbz2`            |
| `tgz`             |
| `txz`             |
| `zip`             |

### `-z`, `--gzip`

Compress the archive with gzip. This is automatically set if the format is
`tar.gz` or `tgz` and is ignored if the format is `zip`.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

Recreate config file from template.

### `-P`, `--parent-dirs`

Also perform command on all parent directories of *target*.

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

## Examples

```console
$ chezmoi archive | tar tvf -
$ chezmoi archive --output=dotfiles.tar.gz
$ chezmoi archive --output=dotfiles.zip
```

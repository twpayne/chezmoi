# `unmanaged` [*path*...]

List all unmanaged files in *path*s. When no *path*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *path*s that are not found on the filesystem.

## Common flags

### `-p`, `--path-style` *style*

--8<-- "common-flags/path-style.md:no-source-tree"

### `-t`, `--tree`

--8<-- "common-flags/tree.md"

## Examples

```console
$ chezmoi unmanaged
$ chezmoi unmanaged ~/.config/chezmoi ~/.ssh
```

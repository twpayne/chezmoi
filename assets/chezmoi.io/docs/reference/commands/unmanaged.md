# `unmanaged` [*path*...]

List all unmanaged files in *path*s. When no *path*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *path*s that are not found on the file system.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `-0`, `--nul-path-separator`

--8<-- "common-flags/nul-path-separator.md"

### `-p`, `--path-style` *style*

--8<-- "common-flags/path-style.md:no-source-tree"

### `-t`, `--tree`

--8<-- "common-flags/tree.md"

## Examples

```sh
chezmoi unmanaged
chezmoi unmanaged ~/.config/chezmoi ~/.ssh
```

# `unmanaged` [*path*...]

List all unmanaged files in *path*s. When no *path*s are supplied, list all
unmanaged files in the destination directory.

It is an error to supply *path*s that are not found on the filesystem.

## Common flags

### `-p`, `--path-style` `absolute`|`relative`|`source-absolute`|`source-relative`

Print paths in the given style. Relative paths are relative to the destination
directory. The default is `relative`.

### `-t`, `--tree`

Print paths as a tree.

## Examples

```console
$ chezmoi unmanaged
$ chezmoi unmanaged ~/.config/chezmoi ~/.ssh
```

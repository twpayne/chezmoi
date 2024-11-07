# `managed` [*path*...]

List all managed entries in the destination directory under all *path*s in
alphabetical order. When no *path*s are supplied, list all managed entries in
the destination directory in alphabetical order.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--format` `json`|`yaml`

--8<-- "common-flags/format.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `-p`, `--path-style` *style*

--8<-- "common-flags/path-style.md:all"

### `-t`, `--tree`

--8<-- "common-flags/tree.md"

## Examples

```console
$ chezmoi managed
$ chezmoi managed --include=files
$ chezmoi managed --include=files,symlinks
$ chezmoi managed -i dirs
$ chezmoi managed -i dirs,files
$ chezmoi managed -i files ~/.config
$ chezmoi managed --exclude=encrypted --path-style=source-relative
```

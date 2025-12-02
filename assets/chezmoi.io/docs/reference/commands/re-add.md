# `re-add` [*target*...]

Re-add modified files in the target state, preserving any `encrypted_`
attributes. chezmoi will not overwrite templates, and all entries that are not
files are ignored. Directories are recursed into by default.

If no *target*s are specified then all modified files are re-added. If one or
more *target*s are given then only those targets are re-added.

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--re-encrypt`

Re-encrypt encrypted files.

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-true"

## Examples

```sh
chezmoi re-add
chezmoi re-add ~/.bashrc
chezmoi re-add --recursive=false ~/.config/git
```

## Notes

!!! hint

    If you want to re-add a single file unconditionally, use `chezmoi add --force` instead.

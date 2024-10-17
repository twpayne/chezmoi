# `archive` [*target*....]

Generate an archive of the target state, or only the targets specified. This
can be piped into `tar` to inspect the target state.

## `-x`, `--exclude` *types*

Exclude entries of type [*types*](../command-line-flags/common.md#-i-include-types),  defaults to `none`.

## `-f`, `--format` *format*

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

## `-i`, `--include` *types*

Only add entries of type [*types*](../command-line-flags/common.md#available-types), defaults to `all`.

## `--init`

Recreate config file from template.

## `-z`, `--gzip`

Compress the archive with gzip. This is automatically set if the format is
`tar.gz` or `tgz` and is ignored if the format is `zip`.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default. Can be disabled with `--recursive=false`.

!!! example

    ```console
    $ chezmoi archive | tar tvf -
    $ chezmoi archive --output=dotfiles.tar.gz
    $ chezmoi archive --output=dotfiles.zip
    ```

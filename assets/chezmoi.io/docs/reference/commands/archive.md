# `archive` [*target*....]

Generate an archive of the target state, or only the targets specified. This
can be piped into `tar` to inspect the target state.

## `-f`, `--format` `tar`|`tar.gz`|`tgz`|`zip`

Write the archive in *format*. If `--output` is set the format is guessed from
the extension, otherwise the default is `tar`.

## `-i`, `--include` *types*

Only include entries of type *types*.

## `-z`, `--gzip`

Compress the archive with gzip. This is automatically set if the format is
`tar.gz` or `tgz` and is ignored if the format is `zip`.

!!! example

    ```console
    $ chezmoi archive | tar tvf -
    $ chezmoi archive --output=dotfiles.tar.gz
    $ chezmoi archive --output=dotfiles.zip
    ```

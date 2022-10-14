# Common command line flags

The following flags apply to multiple commands where they are relevant.

## `-f`, `--format` `json`|`yaml`

Set the output format.

## `-i`, `--include` *types*

Only operate on target state entries of type *types*. *types* is a
comma-separated list of target states (`all`, `dirs`, `files`, `remove`,
`scripts`, `symlinks`) and/or source attributes (`encrypted`, `externals`,
`templates`) and can be excluded by preceding them with a `no`.

!!! example

    `--include=dirs,files` will cause the command to apply to directories and
    files only.

## `--init`

Regenerate and reread the config file from the config file template before
computing the target state.

## `--interactive`

Prompt before applying each target.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default.

## `-x`, `--exclude` *types*

Exclude target state entries of type *types*. *types* is a comma-separated list
of target states (`all`, `dirs`, `files`, `remove`, `scripts`, `symlinks`)
and/or source attributes (`encrypted`, `externals`, `templates`).

!!! example

    `--exclude=scripts` will cause the command to not run scripts and
    `--exclude=encrypted` will exclude encrypted files.

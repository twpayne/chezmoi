# Global command line flags

## `--cache` *directory*

> Configuration: `cacheDir`

Use *directory* as the cache directory.

## `--color` *value*

> Configuration: `color`

Colorize diffs, *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `promptBool`. The default is `auto` which will colorize diffs only
if the environment variable `$NO_COLOR` is not set and stdout is a terminal.

## `-c`, `--config` *filename*

Read the [configuration](/reference/configuration-file) from *filename*.

## `--config-format` `json`|`jsonc`|`toml`|`yaml`

Assume the configuration file is in the given format. This is only needed if
the config filename does not have an extension, for example when it is
`/dev/stdin`.

## `-D`, `--destination` *directory*

> Configuration: `destDir`

Use *directory* as the destination directory.

## `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print
changes that would be made without making them.

## `--force`

Make changes without prompting.

## `-h`, `--help`

Print help.

## `-k`, `--keep-going`

Keep going as far as possible after a encountering an error.

## `--no-pager`

Do not use the pager.

## `--no-tty`

Do not attempt to get a TTY for prompts. Instead, read them from stdin.

## `-o`, `--output` *filename*

Write the output to *filename* instead of stdout.

## `--persistent-state` *filename*

> Configuration: `persistentState`

Read and write the persistent state from *filename*. By default, chezmoi stores
its persistent state in `chezmoistate.boltdb` in the same directory as its
configuration file.

## `--progress` *value*

Show progress when downloading externals. *value* can be `on`, `off`, or `auto`.
The default is `auto` which shows progress bars when stdout is a terminal.

## `-R`, `--refresh-externals` [*value*]

Control the refresh of the externals cache. *value* can be any of `always`,
`auto`, or `never` and defaults to `always` if no *value* is specified. If no
`--refresh-externals` flag is specified then chezmoi defaults to `auto`.

`always` (or any truthy value as accepted by `parseBool`) causes chezmoi to
re-download externals.

`auto` means only re-download externals that have not been downloaded within
their refresh periods.

`never` (or any other falsey value accepted by `parseBool`) means only download
if no cached external is available.

## `-S`, `--source` *directory*

> Configuration: `sourceDir`

Use *directory* as the source directory.

## `--use-builtin-age` *value*

> Configuration: `useBuiltinAge`

Use chezmoi's builtin [age encryption](https://age-encryption.org) instead of an
external `age` command. *value* can be `on`, `off`, `auto`, or any boolean-like
value recognized by `promptBool`. The default is `auto` which will only use the
builtin age if `age.command` cannot be found in `$PATH`.

The builtin `age` command does not support passphrases, symmetric encryption,
or the use of SSH keys.

## `--use-builtin-git` *value*

> Configuration: `useBuiltinGit`

Use chezmoi's builtin git instead of `git.command` for the `init` and `update`
commands. *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `promptBool`. The default is `auto` which will only use the
builtin git if `git.command` cannot be found in `$PATH`.

!!! info

    chezmoi's builtin git has only supports the HTTP and HTTPS transports and
    does not support `git-repo` externals.

## `-v`, `--verbose`

Set verbose mode. In verbose mode, chezmoi prints the changes that it is making
as approximate shell commands, and any differences in files between the target
state and the destination set are printed as unified diffs.

## `--version`

Print the version of chezmoi, the commit at which it was built, and the build
timestamp.

## `-w`, `--working-tree` *directory*

Use *directory* as the git working tree directory. By default, chezmoi searches
the source directory and then its ancestors for the first directory that
contains a `.git` directory.


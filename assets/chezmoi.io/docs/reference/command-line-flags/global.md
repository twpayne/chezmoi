# Global command line flags

The following flags are available for all chezmoi commands. Note that some flags
may not have any effect on certain commands.

## Flags

### `--cache` *directory*

> Configuration: `cacheDir`

Use *directory* as the cache directory.

### `--backup` *archive*

> Configuration: `backup`

Back up replaced files that differ from the target to *archive*. The archive is
created if it does not exist, and subsequent runs append to it.

### `--color` *value*

> Configuration: `color`

Colorize diffs, *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `promptBool`. The default is `auto` which will colorize diffs only
if the environment variable `$NO_COLOR` is not set and stdout is a terminal.

### `-c`, `--config` *filename*

Read the [configuration][configuration] from *filename*.

### `--config-format` *format*

Assume the configuration file is in the given format. This is only needed if the
config filename does not have an extension, for example when it is `/dev/stdin`.
Supported formats: `json`, `jsonc`, `toml`, `yaml`.

### `-D`, `--destination` *directory*

> Configuration: `destDir`

Use *directory* as the destination directory.

### `-n`, `--dry-run`

Set dry run mode. In dry run mode, the destination directory is never modified.
This is most useful in combination with the `-v` (verbose) flag to print changes
that would be made without making them.

### `--force`

Make changes without prompting.

### `--interactive`

Prompt before applying each target.

### `-k`, `--keep-going`

Keep going as far as possible after a encountering an error.

### `--mode` `file`|`symlink`

Mode of operation. The default is `file`.

### `--no-pager`

Do not use the pager.

### `--no-tty`

Do not attempt to get a TTY for prompts. Instead, read them from stdin.

### `-o`, `--output` *filename*

Write the output to *filename* instead of stdout.

### `--persistent-state` *filename*

> Configuration: `persistentState`

Read and write the persistent state from *filename*. By default, chezmoi stores
its persistent state in `chezmoistate.boltdb` in the same directory as its
configuration file.

### `--progress` *value*

Show progress when downloading externals. *value* can be `on`, `off`, or `auto`.
The default is `auto` which shows progress bars when stdout is a terminal.

### `-R`, `--refresh-externals` [*value*]

Control the refresh of the externals cache. *value* can be any of `always`,
`auto`, or `never` and defaults to `always` if no *value* is specified. If no
`--refresh-externals` flag is specified then chezmoi defaults to `auto`.

`always` (or any truthy value as accepted by `parseBool`) causes chezmoi to
re-download externals.

`auto` means only re-download externals that have not been downloaded within
their refresh periods.

`never` (or any other falsy value accepted by `parseBool`) means only download
if no cached external is available.

### `-S`, `--source` *directory*

> Configuration: `sourceDir`

Use *directory* as the source directory.

### `--source-path`

Interpret *targets* passed to the command as paths in the source directory
rather than the destination directory.

### `--use-builtin-age` [*bool*]

> Configuration: `useBuiltinAge`

Use chezmoi's builtin [age encryption][age] instead of an external `age`
command. *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `promptBool`. The default is `auto` which will only use the
builtin age if `age.command` cannot be found in `$PATH`.

The builtin `age` command does not support passphrases, symmetric encryption,
or the use of SSH keys.

### `--use-builtin-diff` [*bool*]

Use chezmoi's builtin diff, even if the `diff.command` configuration variable
is set.

### `--use-builtin-git` [*bool*]

> Configuration: `useBuiltinGit`

Use chezmoi's builtin git instead of `git.command` for the `init` and `update`
commands. *value* can be `on`, `off`, `auto`, or any boolean-like value
recognized by `promptBool`. The default is `auto` which will only use the
builtin git if `git.command` cannot be found in `$PATH`.

!!! info

    chezmoi's builtin git has only supports the HTTP and HTTPS transports and
    does not support `git-repo` externals.

### `-v`, `--verbose`

Set verbose mode. In verbose mode, chezmoi prints the changes that it is making
as approximate shell commands, and any differences in files between the target
state and the destination set are printed as unified diffs.

### `--version`

Print the version of chezmoi, the commit at which it was built, and the build
timestamp.

### `-w`, `--working-tree` *directory*

Use *directory* as the git working tree directory. By default, chezmoi searches
the source directory and then its ancestors for the first directory that
contains a `.git` directory.

[configuration]: /reference/configuration-file/index.md
[age]: https://age-encryption.org

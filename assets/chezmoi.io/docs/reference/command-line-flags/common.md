# Common command line flags

The following flags apply to multiple commands where they are relevant.

## `-f`, `--format` `json`|`yaml`

Set the output format.

## `-i`, `--include` *types*

Include target state entries of type *types*. *types* is a comma-separated list
of types:

| Type        | Description                 |
| ----------- | --------------------------- |
| `all`       | All entries (default)       |
| `none`      | No entries                  |
| `dirs`      | Directories                 |
| `files`     | Files                       |
| `remove`    | Removes                     |
| `scripts`   | Scripts                     |
| `symlinks`  | Symbolic links              |
| `always`    | Scripts that are always run |
| `encrypted` | Encrypted entries           |
| `externals` | External entries            |
| `templates` | Templates                   |

Types can be preceded with `no` to remove them.

Types can be explicitly excluded with the `--exclude` flag.

!!! example

    `--include=files` specifies all files.

## `--init`

Regenerate and reread the config file from the config file template before
computing the target state.

## `--interactive`

Prompt before applying each target.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default.

## `--source-path`

Interpret *targets* passed to the command as paths in the source directory
rather than the destination directory.

## `-x`, `--exclude` *types*

Exclude target state entries of type *types*. *types* is defined as in the
`--include` flag and defaults to `none`.

!!! example

    `--exclude=scripts` will cause the command to not run scripts and
    `--exclude=encrypted` will exclude encrypted files.

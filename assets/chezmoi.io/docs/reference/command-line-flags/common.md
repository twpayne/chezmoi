# Common command line flags

The following flags apply to multiple commands where they are relevant.

## `-x`, `--exclude` *types*

Exclude target state entries of type [*types*](#available-types). Defaults to `none`.

!!! example

    `--exclude=scripts` will cause the command to not run scripts and
    `--exclude=encrypted` will exclude encrypted files.

## `-f`, `--format` `json`|`yaml`

Set the output format.

## `-h`, `--help`

Print help.

## `-i`, `--include` *types*

Include target state entries of type *types*.

!!! example

    `--include=files` specifies all files.

### Available types

*types* is a comma-separated list of types:

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

## `--init`

Regenerate and reread the config file from the config file template before
computing the target state.

## `-P`, `--parent-dirs`

Also perform command on all parent directories of *target*.

## `-p`, `--path-style` `absolute`|`relative`|`source-absolute`|`source-relative`

Print paths in the given style. Relative paths are relative to the destination
directory. The default is `relative`.

## `-r`, `--recursive`

Recurse into subdirectories, `true` by default.

## `--tree`

Print paths as a tree instead of a list.

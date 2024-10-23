# Common command line flags

The following flags apply to multiple commands where they are relevant.

## Flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--format` `json`|`yaml`

Set the output format.

### `-h`, `--help`

Print help.

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

Regenerate and reread the config file from the config file template before
computing the target state.

### `-P`, `--parent-dirs`

Also perform command on all parent directories of *target*.

### `-p`, `--path-style` *style*

Print paths in the given style. The default is `relative`.

| Style             | Description                                 |
| ----------------- | ------------------------------------------- |
| `absolute`        | Absolute paths in the destination directory |
| `relative`        | Relative paths to the destination directory |
| `source-absolute` | Absolute paths in the source tree directory |
| `source-relative` | Relative paths to the source tree directory |

### `-r`, `--recursive`

Recurse into subdirectories, `true` by default.

### `--tree`

Print paths as a tree instead of a list.

## Available entry types

You can provide a list of entry types, separated by commas.
Types can be preceded with `no` to remove them, e.g. `scripts,noalways`.

| Type        | Description                 |
| ----------- | --------------------------- |
| `all`       | All entries                 |
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

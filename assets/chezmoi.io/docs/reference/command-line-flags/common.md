# Common command line flags

The following flags apply to multiple commands where they are relevant.

## Flags

### `--age-recipient` *recipient*

Temporarily override the age recipient for this command. This
only has an effect if age encryption is configured and the `--encrypt` flag is
passed, and cannot be combined with `--age-recipient-file`.

### `--age-recipient-file` *recipient-file*

Temporarily override the age recipient for this command. This
only has an effect if age encryption is configured and the `--encrypt` flag is
passed, and cannot be combined with `--age-recipient`.

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-f`, `--format` `json`|`yaml`

--8<-- "common-flags/format.md"

### `-h`, `--help`

Print help.

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

--8<-- "common-flags/init.md"

### `--override-data` *json-data*

--8<-- "common-flags/override-data.md"

### `-P`, `--parent-dirs`

--8<-- "common-flags/parent-dirs.md"

### `-p`, `--path-style` *style*

--8<-- "common-flags/path-style.md:all"

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-false"

### `--tree`

--8<-- "common-flags/tree.md"

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

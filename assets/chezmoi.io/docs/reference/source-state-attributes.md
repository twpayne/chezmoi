# Source state attributes

chezmoi stores the source state of files, symbolic links, and directories in
regular files and directories in the source directory (`~/.local/share/chezmoi`
by default). This location can be overridden with the `-S` flag or by giving a
value for `sourceDir` in the configuration file. Directory targets are
represented as directories in the source state. All other target types are
represented as files in the source state. Some state is encoded in the source
names.

The following prefixes and suffixes are special, and are collectively referred
to as "attributes":

| Prefix        | Effect                                                                              |
| ------------- | ----------------------------------------------------------------------------------- |
| `after_`      | Run script after updating the destination                                           |
| `before_`     | Run script before updating the destination                                          |
| `create_`     | Ensure that the file exists, and create it with contents if it does not             |
| `dot_`        | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`                          |
| `empty_`      | Ensure the file exists, even if is empty. By default, empty files are removed       |
| `encrypted_`  | Encrypt the file in the source state                                                |
| `external_`   | Ignore attributes in child entries                                                  |
| `exact_`      | Remove anything not managed by chezmoi                                              |
| `executable_` | Add executable permissions to the target file                                       |
| `literal_`    | Stop parsing prefix attributes                                                      |
| `modify_`     | Treat the contents as a script that modifies an existing file                       |
| `once_`       | Only run the script if its contents have not been run before                        |
| `onchange_`   | Only run the script if its contents have not been run before with the same filename |
| `private_`    | Remove all group and world permissions from the target file or directory            |
| `readonly_`   | Remove all write permissions from the target file or directory                      |
| `remove_`     | Remove the file or symlink if it exists or the directory if it is empty             |
| `run_`        | Treat the contents as a script to run                                               |
| `symlink_`    | Create a symlink instead of a regular file                                          |

| Suffix     | Effect                                              |
| ---------- | --------------------------------------------------- |
| `.literal` | Stop parsing suffix attributes                      |
| `.tmpl`    | Treat the contents of the source file as a template |

Different target types allow different prefixes and suffixes. The order of
prefixes is important.

| Target type      | Source type | Allowed prefixes in order                                                         | Allowed suffixes |
| ---------------- | ----------- | --------------------------------------------------------------------------------- | ---------------- |
| Directory        | Directory   | `external_`, `exact_`, `private_`, `readonly_`, `dot_`                            | *none*           |
| Remove directory | Directory   | `remove_`, `dot_`                                                                 | *none*           |
| Regular file     | File        | `encrypted_`, `private_`, `executable_`, `empty_`, `dot_`                         | `.tmpl`          |
| Create file      | File        | `create_`, `encrypted_`, `private_`, `readonly_`, `empty_`, `executable_`, `dot_` | `.tmpl`          |
| Modify file      | File        | `modify_`, `encrypted_`, `private_`, `readonly_`, `executable_`, `dot_`           | `.tmpl`          |
| Remove file      | File        | `remove_`, `dot_`                                                                 | *none*           |
| Script           | File        | `run_`, `once_` or `onchange_`, `before_` or `after_`                             | `.tmpl`          |
| Symbolic link    | File        | `symlink_`, `dot_`                                                                | `.tmpl`          |

The `literal_` prefix and `.literal` suffix can appear anywhere and stop
attribute parsing. This permits filenames that would otherwise conflict with
chezmoi's attributes to be represented.

In addition, if the source file is encrypted, the suffix `.age` (when age
encryption is used) or `.asc` (when gpg encryption is used) is stripped. These
suffixes can be overridden with the `age.suffix` and `gpg.suffix` configuration
variables.

chezmoi ignores all files and directories in the source directory that begin
with a `.` with the exception of files and directories that begin with
`.chezmoi`.

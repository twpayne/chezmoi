# Explanation

## How chezmoi stores its state

For an example of how chezmoi stores its state, see
[`github.com/twpayne/dotfiles`](https://github.com/twpayne/dotfiles).

chezmoi stores the desired state of files, symbolic links, and directories in
regular files and directories in `~/.local/share/chezmoi`. This location can be
overridden with the `-S` flag or by giving a value for `sourceDir` in
`~/.config/chezmoi/chezmoi.toml`.  Some state is encoded in the source names.
chezmoi ignores all files and directories in the source directory that begin
with a `.`. The following prefixes and suffixes are special, and are
collectively referred to as "attributes":

| Prefix/suffix        | Effect                                                                            |
| -------------------- | ----------------------------------------------------------------------------------|
| `encrypted_` prefix  | Encrypt the file in the source state.                                             |
| `private_` prefix    | Remove all group and world permissions from the target file or directory.         |
| `empty_` prefix      | Ensure the file exists, even if is empty. By default, empty files are removed.    |
| `exact_` prefix      | Remove anything not managed by chezmoi.                                           |
| `executable_` prefix | Add executable permissions to the target file.                                    |
| `symlink_` prefix    | Create a symlink instead of a regular file.                                       |
| `dot_` prefix        | Rename to use a leading dot, e.g. `dot_foo` becomes `.foo`.                       |
| `.tmpl` suffix       | Treat the contents of the source file as a template.                              |

Order is important, the order is `exact_`, `private_`, `empty_`, `executable_`,
`symlink_`, `dot_`, `.tmpl`.

Different target types allow different prefixes and suffixes:

| Target type   | Allowed prefixes and suffixes                                      |
| ------------- | ------------------------------------------------------------------ |
| Directory     | `exact_`, `private_`, `dot_`                                       |
| Regular file  | `encrypted_`, `private_`, `empty_`, `executable_`, `dot_`, `.tmpl` |
| Symbolic link | `symlink_`, `dot_`, `.tmpl`                                        |

You can change the attributes of a target in the source state with the `chattr`
command. For example, to make `~/.netrc` private and a template:

    chezmoi chattr private,template ~/.netrc

This only updates the source state of `~/.netrc`, you will need to run `apply`
to apply the changes to the destination state:

    chezmoi apply ~/.netrc

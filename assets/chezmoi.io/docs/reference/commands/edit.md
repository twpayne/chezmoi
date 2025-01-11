# `edit` [*target*...]

Edit the source state of *target*s, which must be files or symlinks. If no
targets are given then the working tree of the source directory is opened.

Encrypted files are decrypted to a private temporary directory and the editor
is invoked with the decrypted file. When the editor exits the edited decrypted
file is re-encrypted and replaces the original file in the source state.

If the operating system supports hard links, then the edit command invokes the
editor with filenames which match the target filename, unless the
`edit.hardlink` configuration variable is set to `false` or the
`--hardlink=false` command line flag is set.

## Flags

### `-a`, `--apply`

> Configuration: `edit.apply`

Apply target immediately after editing. Ignored if there are no targets.

### `--hardlink` *bool*

> Configuration: `edit.hardlink`

Invoke the editor with a hard link to the source file with a name matching the
target filename. This can help the editor determine the type of the file
correctly. This is the default.

### `--watch`

> Configuration: `edit.watch`

Automatically apply changes when files are saved, with the following limitations:

* Only available when `chezmoi edit` is invoked with arguments (i.e.
  argument-free `chezmoi edit` is not supported).
* All edited files are applied when any file is saved.
* Only the edited files are watched, not any dependent files (e.g.
  `.chezmoitemplates` and `include`d files in templates are not watched).
* Only works on operating systems supported by
  [fsnotify](https://github.com/fsnotify/fsnotify).

## Common flags

### `-x`, `--exclude` *types*

--8<-- "common-flags/exclude.md"

### `-i`, `--include` *types*

--8<-- "common-flags/include.md"

### `--init`

--8<-- "common-flags/init.md"

## Examples

  ```sh
  chezmoi edit ~/.bashrc
  chezmoi edit ~/.bashrc --apply
  chezmoi edit
  ```

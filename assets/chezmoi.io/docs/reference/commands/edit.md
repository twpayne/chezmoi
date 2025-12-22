# `edit` [*target*...]

Edit the source state of *target*s, which must be files or symlinks. If no
targets are given then the working tree of the source directory is opened.

Encrypted files are decrypted to a private temporary directory and the editor
is invoked with the decrypted file. When the editor exits the edited decrypted
file is re-encrypted and replaces the original file in the source state.

If the operating system supports hard links, then the edit command invokes the
editor with filenames which match the target filename, unless the
`edit.hardlink` configuration variable is set to `false` or the
`--hardlink=false` command line flag is set. Templates preserve their `.tmpl`
extension so editors can highlight them as templates.

!!! hint

    Depending on your editor, you can set the format of a file in the file itself
    using a [modeline][modelines]. This can be useful if you want to syntax
    highlight a template as a different format.

## Flags

### `-a`, `--apply`

> Configuration: `edit.apply`

Apply target immediately after editing. Ignored if there are no targets.

### `--hardlink` *bool*

> Configuration: `edit.hardlink`

Invoke the editor with a hard link to the source file with a name matching the
target filename. This can help the editor determine the type of the file
correctly. This is the default.

!!! hint

    Creating hardlinks is not possible between different filesystems. Hence,
    if your [`tempDir`][tempdir] resides on a different filesystem (e.g. a
    [tmpfs][tmpfs], which is sometimes used for `/tmp`), this will not work.

### `--watch`

> Configuration: `edit.watch`

Automatically apply changes when files are saved, with the following limitations:

* Only available when `chezmoi edit` is invoked with arguments (i.e.
  argument-free `chezmoi edit` is not supported).
* All edited files are applied when any file is saved.
* Only the edited files are watched, not any dependent files (e.g.
  `.chezmoitemplates` and `include`d files in templates are not watched).
* Only works on operating systems supported by [fsnotify][fsnotify].
* Only works if `edit.hardlink` is enabled and works.

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

[fsnotify]: https://github.com/fsnotify/fsnotify
[modelines]: https://vimhelp.org/options.txt.html#auto-setting
[tempdir]: /reference/configuration-file/variables.md#tempdir
[tmpfs]: https://en.wikipedia.org/wiki/Tmpfs

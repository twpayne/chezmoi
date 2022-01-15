# `edit` [*target*...]

Edit the source state of *target*s, which must be files or symlinks. If no
targets are given then the working tree of the source directory is opened.

Encrypted files are decrypted to a private temporary directory and the editor
is invoked with the decrypted file. When the editor exits the edited decrypted
file is re-encrypted and replaces the original file in the source state.

If the operating system supports hard links, then the edit command invokes the
editor with filenames which match the target filename, unless the
`edit.hardlink` configuration variable is set to `false` the `--hardlink=false`
command line flag is set.

## `-a`, `--apply`

Apply target immediately after editing. Ignored if there are no targets.

## `--hardlink` *bool*

Invoke the editor with a hard link to the source file with a name matching the
target filename. This can help the editor determine the type of the file
correctly. This is the default.

!!! example

    ```console
    $ chezmoi edit ~/.bashrc
    $ chezmoi edit ~/.bashrc --apply
    $ chezmoi edit
    ```

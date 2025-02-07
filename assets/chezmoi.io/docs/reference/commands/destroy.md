# `destroy` *target*...

!!! danger

    The `destroy` command permanently removes files both from your home
    directory and chezmoi's source directory.

    Only run `chezmoi destroy` if you have a separate backup of your home
    directory and your source directory.

    If you want chezmoi to stop managing the file use [`forget`][forget]
    instead.

    If you want to remove all traces of chezmoi from your system use
    [`purge`][purge] instead.

Remove *target* from the source state, the destination directory, and the state.

## Common flags

### `--force`

Destroy without prompting.

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-false"

[forget]: /reference/commands/forget.md
[purge]: /reference/commands/purge.md

# `.chezmoiroot`

If a file called `.chezmoiroot` exists in the root of the source directory then
the source state is read from the directory specified in `.chezmoiroot`
interpreted as a relative path to the source directory. `.chezmoiroot` is read
before all other files in the source directory.

!!! warning

    If you use this approach, you must move all other "source root" files such
    as [`.chezmoi.$FORMAT.tmpl`][config] into your new root.

[config]: /reference/special-files/chezmoi-format-tmpl.md
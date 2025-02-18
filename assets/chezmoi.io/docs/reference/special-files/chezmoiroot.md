# `.chezmoiroot`

If a file called `.chezmoiroot` exists in the root of the source directory then
the source state is read from the directory specified in `.chezmoiroot`
interpreted as a relative path to the source directory. `.chezmoiroot` is read
after [`.chezmoiversion`][version] but before all other files in the source
directory.

[version]: /reference/special-files/chezmoiversion.md

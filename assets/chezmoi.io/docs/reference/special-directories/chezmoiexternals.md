# `.chezmoiexternals/`

If any `.chezmoiexternals/` directories exist in the source state, then all
files in this directory are treated as [`.chezmoiexternal.<format>`][external]
files relative to the source directory.

!!! warning

    `.chezmoiexternals/` directories do not support externals for subdirectories
    within the `.chezmoiexternals/` directories. See [#4274][issue-4274] for
    details.

[external]: /reference/special-files/chezmoiexternal-format.md
[issue-4274]: https://github.com/twpayne/chezmoi/issues/4274

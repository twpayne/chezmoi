# `.chezmoiversion`

If a file called `.chezmoiversion` exists anywhere in the source directory (not
just the source state), then its contents are interpreted as a semantic version
defining the minimum version of chezmoi required to interpret the source state
correctly. chezmoi will refuse to interpret the source state if the current
version is too old.

This file is evaluated before any operation.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoiversion"
    2.50.0
    ```

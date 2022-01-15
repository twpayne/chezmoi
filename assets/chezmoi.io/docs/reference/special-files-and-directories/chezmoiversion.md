# `.chezmoiversion`

If a file called `.chezmoiversion` exists, then its contents are interpreted as
a semantic version defining the minimum version of chezmoi required to
interpret the source state correctly. chezmoi will refuse to interpret the
source state if the current version is too old.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoiversion"
    1.5.0
    ```

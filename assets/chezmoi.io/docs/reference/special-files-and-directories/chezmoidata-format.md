# `.chezmoidata.<format>`

If a file called `.chezmoidata.<format>` exists in the source state, it is
interpreted as template data.

!!! example

    If `.chezmoidata.toml` contains the following:

    ```toml title="~/.local/share/chezmoi/.chezmoidata.toml"
    fontSize = 12
    ```

    Then the `.fontSize` variable is available in templates, e.g.

    ```
    FONT_SIZE={{ .fontSize }}
    ```

    Will result in:

    ```
    FONT_SIZE=12
    ```

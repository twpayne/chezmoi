# `.chezmoidata.<format>`

If a file called `.chezmoidata.<format>` exists in the source state, it is
interpreted as a datasource available in most templates.

!!! example

    If `.chezmoidata.toml` contains the following (and no variable is
    overwritten in later stages):

    ```toml title="~/.local/share/chezmoi/.chezmoidata.toml"
    editor = "nvim"
    [directions]
      up = "k"
      down = "j"
      right = "l"
      left = "h"
    ```

    Then the following template:

    ```
    EDITOR={{ .editor }}
    MOVE_UP={{ .directions.up }}
    MOVE_DOWN={{ .directions.down }}
    MOVE_RIGHT={{ .directions.right }}
    MOVE_LEFT={{ .directions.left }}
    ```

    Will result in:

    ```
    EDITOR=nvim
    MOVE_UP=k
    MOVE_DOWN=j
    MOVE_RIGHT=l
    MOVE_LEFT=h
    ```

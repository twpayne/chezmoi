# `generate` *file*

Generates *file* for use with chezmoi. The currently supported *file*s are:

| File         | Description                                                |
| ------------ | ---------------------------------------------------------- |
| `install.sh` | An install script, suitable for use with Github Codespaces |

!!! example

    ```console
    $ chezmoi generate install.sh > install.sh
    ```

# `generate` *output*

Generates *output* for use with chezmoi. The currently supported *output*s are:

| Output               | Description                                                           |
| -------------------- | --------------------------------------------------------------------- |
| `git-commit-message` | A git commit message, describing the changes to the source directory. |
| `install.sh`         | An install script, suitable for use with GitHub Codespaces            |

!!! example

    ```console
    $ chezmoi generate install.sh > install.sh
    $ chezmoi git commit -m "$(chezmoi generate git-commit-message)"
    ```

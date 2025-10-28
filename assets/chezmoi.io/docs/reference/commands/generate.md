# `generate` *output*

Generates *output* for use with chezmoi. The currently supported *output*s are:

| Output                  | Description                                                                   |
| ----------------------- | ----------------------------------------------------------------------------- |
| `git-commit-message`    | A git commit message, describing the changes to the source directory.         |
| `install.sh`            | An install script, suitable for use with GitHub Codespaces                    |
| `install-init-shell.sh` | A script which installs chezmoi, runs `chezmoi init`, and executes your shell |

## Examples

```sh
chezmoi generate install.sh > install.sh
chezmoi git -- commit -m "$(chezmoi generate git-commit-message)"
chezmoi generate install-init-shell.sh $GITHUB_USERNAME
```

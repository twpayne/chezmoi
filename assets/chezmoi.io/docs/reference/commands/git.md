# `git` [*arg*...]

Run `git` *args* in the working tree (typically the source directory).

!!! note

    Flags in *args* must occur after `--` to prevent chezmoi from interpreting
    them.

## Examples

```sh
chezmoi git add .
chezmoi git add dot_gitconfig
chezmoi git -- commit -m "Add .gitconfig"
```

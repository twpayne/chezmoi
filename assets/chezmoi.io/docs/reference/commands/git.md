# `git` [*arg*...]

Run `git` *arg*s in the working tree (typically the source directory). Note
that flags in *arguments* must occur after `--` to prevent chezmoi from
interpreting them.

!!! example

    ```console
    $ chezmoi git add .
    $ chezmoi git add dot_gitconfig
    $ chezmoi git -- commit -m "Add .gitconfig"
    ```

# pass functions

The `pass` template functions return passwords stored in
[pass](https://www.passwordstore.org/) using the pass CLI (`pass`).

!!! hint

    To use a pass-compatible password manager like
    [passage](https://github.com/FiloSottile/passage), set `pass.command` to
    the name of the binary and use chezmoi's `pass*` template functions as if
    you were using pass.

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [pass]
        command = "passage"
    ```

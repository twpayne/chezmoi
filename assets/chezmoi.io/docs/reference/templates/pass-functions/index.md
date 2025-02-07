# pass functions

The `pass` template functions return passwords stored in [pass][pass] using the
pass CLI (`pass`).

!!! hint

    To use a pass-compatible password manager like [passage][passage], set
    `pass.command` to the name of the binary and use chezmoi's `pass*` template
    functions as if you were using pass.

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [pass]
        command = "passage"
    ```

[pass]: https://www.passwordstore.org/
[passage]: https://github.com/FiloSottile/passage,

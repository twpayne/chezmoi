# `lastpassRaw` *id*

`lastpassRaw` returns structured data from [LastPass][lastpass] using the
[LastPass CLI][cli] (`lpass`). It behaves identically to the `lastpass`
function, except that no further parsing is done on the `note` field.

!!! example

    ```
    {{ (index (lastpassRaw "SSH Private Key") 0).note }}
    ```

[lastpass]: https://lastpass.com/
[cli]: https://lastpass.github.io/lastpass-cli/lpass.1.html

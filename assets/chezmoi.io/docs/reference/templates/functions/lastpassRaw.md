# `lastpassRaw` *id*

`lastpassRaw` returns structured data from [LastPass](https://lastpass.com/)
using the [LastPass CLI](https://lastpass.github.io/lastpass-cli/lpass.1.html)
(`lpass`). It behaves identically to the `lastpass` function, except that no
further parsing is done on the `note` field.

!!! example

    ```
    {{ (index (lastpassRaw "SSH Private Key") 0).note }}
    ```

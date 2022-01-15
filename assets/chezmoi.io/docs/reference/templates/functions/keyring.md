# `keyring` *service* *user*

`keyring` retrieves the value associated with *service* and *user* from the
user's keyring.

| OS      | Keyring                     |
| ------- | --------------------------- |
| macOS   | Keychain                    |
| Linux   | GNOME Keyring               |
| Windows | Windows Credentials Manager |

!!! example

    ```
    [github]
        user = {{ .github.user | quote }}
        token = {{ keyring "github" .github.user | quote }}
    ```

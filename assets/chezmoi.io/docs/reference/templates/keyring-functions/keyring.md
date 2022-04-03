# `keyring` *service* *user*

`keyring` retrieves the value associated with *service* and *user* from the
user's keyring.

| OS      | Keyring                     |
| ------- | --------------------------- |
| macOS   | Keychain                    |
| Linux   | GNOME Keyring               |
| Windows | Windows Credentials Manager |
| FreeBSD | GNOME Keyring               |

!!! example

    ```
    [github]
        user = {{ .github.user | quote }}
        token = {{ keyring "github" .github.user | quote }}
    ```

!!! warning

    On FreeBSD, the `keyring` template function is only available if chezmoi
    was compiled with cgo enabled. The official release binaries of chezmoi are
    **not** compiled with cgo enabled, and `keyring` will always return an
    empty string.

# `onepasswordDocument` *uuid* [*vault* [*account*]]

`onepasswordDocument` returns a document from [1Password][1p] using the
[1Password CLI][op] (`op`). *uuid* is passed to `op get document $UUID` and the
output from `op` is returned. The output from `op` is cached so calling
`onepasswordDocument` multiple times with the same *uuid* will only invoke `op`
once. If the optional *vault* is supplied, it will be passed along to the `op
get` call, which can significantly improve performance. If the optional
*account* is supplied, it will be passed along to the `op get` call, which will
help it look in the right account, in case you have multiple accounts (e.g.,
personal and work accounts).

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

!!! example

    ```
    {{- onepasswordDocument "$UUID" -}}
    {{- onepasswordDocument "$UUID" "$VAULT_UUID" -}}
    {{- onepasswordDocument "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME" -}}
    {{- onepasswordDocument "$UUID" "" "$ACCOUNT_NAME" -}}
    ```

!!! warning

    When using [1Password Connect][connect], `onepasswordDocument` is not
    available.

!!! warning

    When using [1Password Service Accounts][accounts], the *account* parameter
    is not allowed.

[1p]: https://1password.com/
[op]: https://developer.1password.com/docs/cli
[connect]: /user-guide/password-managers/1password.md#1password-connect
[accounts]: /user-guide/password-managers/1password.md#1password-service-accounts

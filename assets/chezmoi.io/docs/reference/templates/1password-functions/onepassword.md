# `onepassword` *uuid* [*vault-uuid* [*account-name*]]

`onepassword` returns structured data from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`).
*uuid* is passed to `op item get $UUID --format json` and the output from `op`
is parsed as JSON. The output from `op` is cached so calling `onepassword`
multiple times with the same *uuid* will only invoke `op` once. If the optional
*vault-uuid* is supplied, it will be passed along to the `op item get` call,
which can significantly improve performance. If the optional *account-name* is
supplied, it will be passed along to the `op item get` call, which will help it
look in the right account, in case you have multiple accounts (e.g., personal
and work accounts).

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

The 1password CLI command can be set with the `onePassword.command` config
variable, and extra arguments can be specified with the `onePassword.args`
config variable.

!!! example

    ```
    {{ (onepassword "$UUID").fields[1].value }}
    {{ (onepassword "$UUID" "$VAULT_UUID").fields[1].value }}
    {{ (onepassword "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME").fields[1].value }}
    {{ (onepassword "$UUID" "" "$ACCOUNT_NAME").fields[1].value }}
    ```

    A more robust way to get a password field would be something like:

    ```
    {{ range (onepassword "$UUID").fields -}}
    {{   if and (eq .label "password") (eq .purpose "PASSWORD") -}}
    {{     .value -}}
    {{   end -}}
    {{ end }}
    ```

    !!! info

        For 1Password CLI 1.x.

        ```
        {{ (onepassword "$UUID").details.password }}
        {{ (onepassword "$UUID" "$VAULT_UUID").details.password }}
        {{ (onepassword "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME").details.password }}
        {{ (onepassword "$UUID" "" "$ACCOUNT_NAME").details.password }}
        ```

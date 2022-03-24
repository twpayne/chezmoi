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

!!! danger

    When using [1Password CLI 2.0](https://developer.1password.com/), note that
    the structure of the data returned by the `onepassword` template function
    is different and your templates will need updating.

    You may wish to use `onepasswordDetailsFields` or `onepasswordItemFields`
    instead of this function, as `onepassword` returns fields as a list of
    objects. However, this function may return values that are inaccessible from
    the other functions. Testing the output of this function is recommended:

    ```console
    $ chezmoi execute-template "{{ onepassword \"$UUID\" | toJson }}" | jq .
    ```

!!! warning

    When using 1Password CLI 2.0, there may be an issue with pre-authenticating
    `op` because the environment variable used to store the session key has
    changed from `OP_SESSION_account` to `OP_SESSION_accountUUID`. Instead of
    using *account-name*, it is recommended that you use the *account-uuid*.
    This can be found using `op account list`.

    This issue does not occur when using biometric authentication and 1Password
    8, or if you allow chezmoi to prompt you for 1Password authentication
    (`1password.prompt = true`).

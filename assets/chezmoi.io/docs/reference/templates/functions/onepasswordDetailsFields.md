# `onepasswordDetailsFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordDetailsFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item $UUID`, the output from `op` is parsed as JSON, and
elements of `details.fields` are returned as a map indexed by each field's
`designation`. If there is no valid session in the environment, by default you
will be interactively prompted to sign in.

The output from `op` is cached so calling `onepasswordDetailsFields` multiple
times with the same *uuid* will only invoke `op` once. If the optional
*vault-uuid* is supplied, it will be passed along to the `op get` call, which
can significantly improve performance. If the optional *account-name* is
supplied, it will be passed along to the `op get` call, which will help it look
in the right account, in case you have multiple accounts (e.g., personal and
work accounts).

!!! example

    ```
    {{ (onepasswordDetailsFields "$UUID").password.value }}
    {{ (onepasswordDetailsFields "$UUID" "$VAULT_UUID").password.value }}
    {{ (onepasswordDetailsFields "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME").password.value }}
    {{ (onepasswordDetailsFields "$UUID" "" "$ACCOUNT_NAME").password.value }}
    ```

!!! example

    Given the output from `op`:

    ```json
    {
        "uuid": "$UUID",
        "details": {
            "fields": [
                {
                    "designation": "username",
                    "name": "username",
                    "type": "T",
                    "value": "exampleuser"
                },
                {
                    "designation": "password",
                    "name": "password",
                    "type": "P",
                    "value": "examplepassword"
                }
            ]
        }
    }
    ```

    the return value of `onepasswordDetailsFields` will be the map:

    ```json
    {
        "username": {
            "designation": "username",
            "name": "username",
            "type": "T",
            "value": "exampleuser"
        },
        "password": {
            "designation": "password",
            "name": "password",
            "type": "P",
            "value": "examplepassword"
        }
    }
    ```

!!! danger

    When using [1Password CLI 2.0](https://developer.1password.com/), note that
    the structure of the data returned by the `onepasswordDetailsFields`
    template function is different and your templates will need updating.

    You may wish to use `onepassword` or `onepasswordItemFields` instead of this
    function, as it may not return expected values. Testing the output of this
    function is recommended:

    ```console
    chezmoi execute-template "{{- onepasswordDetailsFields \"$UUID\" | toJson -}}" | jq .
    ```

!!! warning

    When using 1Password CLI 2.0, there may be an issue with pre-authenticating
    `op` because the environment variable used to store the session key has
    changed from `OP_SESSION_account` to `OP_SESSION_accountUUID`. Instead of
    using *account-name*, it is recommended that you use the *account-uuid*.
    This can be found using `op account list`.

    This issue does not exist when using biometric authentication and 1Password
    8, or if you allow chezmoi to prompt you for 1Password authentication
    (`1password.prompt = true`).

!!! info

    In earlier versions of chezmoi, if *vault-uuid* or *account-name* were
    empty strings, they would be added to the resulting `op` command-line
    (`--vault ''`). This causes errors in 1Password CLI 2.0, so those arguments
    will no longer be added.

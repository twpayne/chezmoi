# `onepasswordDetailsFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordDetailsFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`).
*uuid* is passed to `op get item $UUID`, the output from `op` is parsed as
JSON, and elements of `details.fields` are returned as a map indexed by each
field's `designation`. If there is no valid session in the environment, by
default you will be interactively prompted to sign in.

The output from `op` is cached so calling `onepasswordDetailsFields` multiple
times with the same *uuid* will only invoke `op` once.  If the optional
*vault-uuid* is supplied, it will be passed along to the `op get` call, which
can significantly improve performance. If the optional *account-name* is
supplied, it will be passed along to the `op get` call, which will help it look
in the right account, in case you have multiple accounts (eg. personal and work
accounts).

!!! example

    ```
    {{ (onepasswordDetailsFields "$UUID").password.value }}
    {{ (onepasswordDetailsFields "$UUID" "$VAULT_UUID").password.value }}
    {{ (onepasswordDetailsFields "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME").password.value }}
    ```

    If using 1Password 1.0, then *vault-uuid* is optional.

    ```
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

!!! info

    If you're using [1Password CLI 2.0](https://developer.1password.com/), there
    are changes to be aware of.

    !!! warning

    The structure of the data returned by the `onepasswordDetailsFields`
    template function will be different and you will need to update your
    templates. The structure has not yet been finalized.

    !!! warning

    Neither *vault-uuid* nor *account-name* may be empty strings if specified.
    Older versions of 1Password CLI would ignore empty strings for arguments.

    !!! warning

    Unless using biometric authentication, or when using without prompting, it
    is recommended that instead of *account-name*, the UUID of the account is
    used. This can be shown with `op account list`.

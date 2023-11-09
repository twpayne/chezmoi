# `onepasswordDetailsFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordDetailsFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *uuid*
is passed to `op get item $UUID`, the output from `op` is parsed as JSON, and
elements of `details.fields` are returned as a map indexed by each field's `id`
(if set) or `label` (if set and `id` is not present).

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

The output from `op` is cached so calling `onepasswordDetailsFields` multiple
times with the same *uuid* will only invoke `op` once. If the optional
*vault-uuid* is supplied, it will be passed along to the `op get` call, which
can significantly improve performance. If the optional _account-name_ is
supplied, it will be passed along to the `op get` call, which will help it look
in the right account, in case you have multiple accounts (e.g. personal and work
accounts).

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

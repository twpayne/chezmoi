# `onepasswordItemFields` *uuid* [*vault* [*account*]]

`onepasswordItemFields` returns structured data from [1Password][1p] using the
[1Password CLI][op] (`op`). *uuid* is passed to `op item get $UUID --format
json`, the output from `op` is parsed as JSON, and each element of
`details.sections` are iterated over and any `fields` are returned as a map
indexed by each field's `n`.

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

!!! example

    The result of

    ```
    {{ (onepasswordItemFields "abcdefghijklmnopqrstuvwxyz").exampleLabel.value }}
    ```

    is equivalent to calling

    ```console
    $ op item get abcdefghijklmnopqrstuvwxyz --fields label=exampleLabel
    ```

    or

    ```console
    $ op item get abcdefghijklmnopqrstuvwxyz --fields exampleLabel
    ```

!!! example

    Given the output from `op`:

    ```json
    {
        "id": "$UUID",
        "title": "$TITLE",
        "version": 1,
        "vault": {
            "id": "$vaultUUID"
        },
        "category": "LOGIN",
        "last_edited_by": "userUUID",
        "created_at": "2022-01-12T16:29:26Z",
        "updated_at": "2022-01-12T16:29:26Z",
        "sections": [
            {
                "id": "$sectionID",
                "label": "Related Items"
            }
        ],
        "fields": [
            {
                "id": "nerlnqbfzdm5q5g6ydsgdqgdw4",
                "type": "STRING",
                "label": "exampleLabel",
                "value": "exampleValue"
            }
        ],
    }
    ```

    the return value of `onepasswordItemFields` will be the map:

    ```json
    {
        "exampleLabel": {
            "id": "string",
            "type": "D4328E0846D2461E8E455D7A07B93397",
            "label": "exampleLabel",
            "value": "exampleValue"
        }
    }
    ```

!!! warning

    When using [1Password secrets automation][automation], the *account*
    parameter is not allowed.

[1p]: https://1password.com/
[op]: https://support.1password.com/command-line-getting-started/
[automation]: /user-guide/password-managers/1password.md#secrets-automation

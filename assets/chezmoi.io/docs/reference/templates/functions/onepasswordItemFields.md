# `onepasswordItemFields` _uuid_ [_vault-uuid_ [*account-name*]]

`onepasswordItemFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). _uuid_
is passed to `op item get $UUID --format json`, the output from `op` is parsed
as JSON, and each element of `details.sections` are iterated over and any
`fields` are returned as a map indexed by each field's `n`.

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
    # or
    $ op item get abcdefghijklmnopqrstuvwxyz --fields exampleLabel
    ```

    ??? info

        For 1Password CLI 1.x.

        ```
        {{ (onepasswordItemFields "abcdefghijklmnopqrstuvwxyz").exampleLabel.v }}
        ```

        is equivalent to calling

        ```console
        $ op get item abcdefghijklmnopqrstuvwxyz --fields exampleLabel
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

    ??? info

        For 1Password CLI 1.x, the output is this:

        ```json
        {
          "uuid": "$UUID",
          "details": {
            "sections": [
              {
                "name": "linked items",
                "title": "Related Items"
              },
              {
                "fields": [
                  {
                    "k": "string",
                    "n": "D4328E0846D2461E8E455D7A07B93397",
                    "t": "exampleLabel",
                    "v": "exampleValue"
                  }
                ],
                "name": "Section_20E0BD380789477D8904F830BFE8A121",
                "title": ""
              }
            ]
          },
        }
        ```

        the return value of `onepasswordItemFields` will be the map:

        ```json
        {
            "exampleLabel": {
                "k": "string",
                "n": "D4328E0846D2461E8E455D7A07B93397",
                "t": "exampleLabel",
                "v": "exampleValue"
            }
        }
        ```

!!! info

    When using [1Password CLI 2.0](https://developer.1password.com/), note that
    the structure of the data returned by the `onepasswordItemFields` template
    function is different and your templates will need updating.

    You may wish to use `onepassword` or `onepasswordDetailsFields` instead of
    this function, as it may not return expected values. Testing the output of
    this function is recommended:

    ```console
    chezmoi execute-template "{{- onepasswordItemFields \"$UUID\" | toJson -}}" | jq .
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

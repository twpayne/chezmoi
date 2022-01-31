# `onepasswordItemFields` *uuid* [*vault-uuid* [*account-name*]]

`onepasswordItemFields` returns structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`).
*uuid* is passed to `op get item <uuid>`, the output from `op` is parsed as
JSON, and each element of `details.sections` are iterated over and any `fields`
are returned as a map indexed by each field's `n`. If there is no valid session
in the environment, by default you will be interactively prompted to sign in.

!!! example

    The result of

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
      "uuid": "<uuid>",
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

    If you're using [1Password CLI 2.0](https://developer.1password.com/), then
    the structure of the data returned by the `onepasswordDetailsFields`
    template function will be different and you will need to update your
    templates.

    !!! warning

    The structure of the data returned will not be finalized until 1Password
    CLI 2.0 is released.

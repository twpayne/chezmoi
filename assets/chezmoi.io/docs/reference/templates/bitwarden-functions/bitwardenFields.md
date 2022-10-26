# `bitwardenFields` [*arg*...]

`bitwardenFields` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://bitwarden.com/help/cli) (`bw`). *arg*s are passed to `bw get`
unchanged, the output from `bw get` is parsed as JSON, and elements of `fields`
are returned as a map indexed by each field's `name`.

The output from `bw get` is cached so calling `bitwarden` multiple times with
the same arguments will only invoke `bw get` once.

!!! example

    ```
    {{ (bitwardenFields "item" "$ITEMID").token.value }}
    ```

!!! example

    Given the output from `bw get`:

    ```json
    {
        "object": "item",
        "id": "bf22e4b4-ae4a-4d1c-8c98-ac620004b628",
        "organizationId": null,
        "folderId": null,
        "type": 1,
        "name": "example.com",
        "notes": null,
        "favorite": false,
        "fields": [
            {
                "name": "hidden",
                "value": "hidden-value",
                "type": 1
            },
            {
                "name": "token",
                "value": "token-value",
                "type": 0
            }
        ],
        "login": {
            "username": "username-value",
            "password": "password-value",
            "totp": null,
            "passwordRevisionDate": null
        },
        "collectionIds": [],
        "revisionDate": "2020-10-28T00:21:02.690Z"
    }
    ```

    the return value if `bitwardenFields` will be the map:

    ```json
    {
        "hidden": {
            "name": "hidden",
            "type": 1,
            "value": "hidden-value"
        },
        "token": {
            "name": "token",
            "type": 0,
            "value": "token-value"
        }
    }
    ```

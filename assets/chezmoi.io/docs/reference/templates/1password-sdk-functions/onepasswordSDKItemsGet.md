# `onepasswordSDKItemsGet` *vault-id* *item-id*

!!! warning

    `onepasswordSDKItemsGet` is an experimental function and may change.

`onepasswordSDKItemsGet` returns an item from [1Password][1p] using the
[1Password SDK][sdk].

The output of `onepasswordSDKItemsGet` is cached so multiple calls to
`onepasswordSDKItemsGet` with the same *vault-id* and *item-id* will return the
same value.

!!! example

    ```
    {{- onepasswordSDKItemsGet "vault" "item" | toJson -}}
    ```

[1p]: https://1password.com
[sdk]: https://developer.1password.com/docs/sdks/

# `onepasswordSDKItemsGet` *vault-id* *item-id*

!!! warning

    `onepasswordSDKItemsGet` is an experimental function and may change.

`onepasswordSDKItemsGet` returns an item from [1Password](https://1password.com)
using the [1Password SDK](https://developer.1password.com/docs/sdks/).

The output of `onepasswordSDKItemsGet` is cached so multiple calls to
`onepasswordSDKItemsGet` with the same *vault-id* and *item-id* will return the same value.

!!! example

    ```
    {{- onepasswordSDKItemsGet "vault" "item" | toJson -}}
    ```

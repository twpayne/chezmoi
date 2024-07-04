# `onepasswordSDKSecretsResolve` *url*

!!! warning

    `onepasswordSDKSecretsResolve` is an experimental function and may change.

`onepasswordSDKSecretsResolve` returns a secret from [1Password](https://1password.com)
using the [1Password SDK](https://developer.1password.com/docs/sdks/).

The output of `onepasswordSDKSecretsResolve` is cached so multiple calls to
`onepasswordSDKSecretsResolve` with the same *url* will return the same value.

!!! example

    ```
    {{- onepasswordSDKSecretsResolve "op://vault/item/field" -}}
    ```

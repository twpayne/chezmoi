# `vault` *key*

`vault` returns structured data from [Vault](https://www.vaultproject.io/)
using the [Vault CLI](https://www.vaultproject.io/docs/commands/) (`vault`).
*key* is passed to `vault kv get -format=json $KEY` and the output from `vault`
is parsed as JSON. The output from `vault` is cached so calling `vault`
multiple times with the same *key* will only invoke `vault` once.

!!! example

    ```
    {{ (vault "$KEY").data.data.password }}
    ```

# `vault` *key*

`vault` returns structured data from [Vault][vault] using the [Vault CLI][cli]
(`vault`). *key* is passed to `vault kv get -format=json $KEY` and the output
from `vault` is parsed as JSON. The output from `vault` is cached so calling
`vault` multiple times with the same *key* will only invoke `vault` once.

!!! example

    ```
    {{ (vault "$KEY").data.data.password }}
    ```

[vault]: https://www.vaultproject.io/
[cli]: https://www.vaultproject.io/docs/commands/

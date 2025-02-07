# Vault

chezmoi includes support for [Vault][vault] using the [Vault CLI][cli] to expose
data as a template function.

The vault CLI needs to be correctly configured on your machine, e.g. the
`VAULT_ADDR` and `VAULT_TOKEN` environment variables must be set correctly.
Verify that this is the case by running:

```sh
vault kv get -format=json $KEY
```

The structured data from `vault kv get -format=json` is available as the `vault`
template function. You can use the `.Field` syntax of the `text/template`
language to extract the data you want. For example:

```text
{{ (vault "$KEY").data.data.password }}
```

[vault]: https://www.vaultproject.io/
[cli]: https://www.vaultproject.io/docs/commands/

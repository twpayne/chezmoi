# Vault

chezmoi includes support for [Vault](https://www.vaultproject.io/) using the
[Vault CLI](https://www.vaultproject.io/docs/commands/) to expose data as a
template function.

The vault CLI needs to be correctly configured on your machine, e.g. the
`VAULT_ADDR` and `VAULT_TOKEN` environment variables must be set correctly.
Verify that this is the case by running:

```console
$ vault kv get -format=json $KEY
```

The structured data from `vault kv get -format=json` is available as the
`vault` template function. You can use the `.Field` syntax of the
`text/template` language to extract the data you want. For example:

```
{{ (vault "$KEY").data.data.password }}
```

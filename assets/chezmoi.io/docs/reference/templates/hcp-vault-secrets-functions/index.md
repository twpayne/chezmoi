# HCP Vault Secrets

chezmoi includes support for [HCP Vault Secrets][secrets] using the `vlt` CLI to
expose data through the `hcpVaultSecret` and `hcpVaultSecretJson` template
functions.

!!! info

    If you access HCP Vault Secrets through the `hcp`, these functions **will
    not work**. See [`vlt` vs `hcp`: Upgrades that Break][break].

[secrets]: https://developer.hashicorp.com/hcp/docs/vault-secrets
[break]: /user-guide/password-managers/hcp-vault-secrets.md#hcp-broken

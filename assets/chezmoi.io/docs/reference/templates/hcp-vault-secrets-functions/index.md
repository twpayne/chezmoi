# HCP Vault Secrets

chezmoi includes support for [HCP Vault Secrets][secrets] using the `vlt` CLI to
expose data through the `hcpVaultSecret` and `hcpVaultSecretJson` template
functions.

!!! info

    If you access HCP Vault Secrets through the `hcp`, these functions **will
    not work**. See [`vlt` vs `hcp`: Upgrades that Break][break].

!!! warning

    Hashicorp has announced the [closure the Vault Secrets service][eol] for
    most customers 27 August 2025 with a recommendation of migrating to
    Hashicorp Vault, which is an enterprise-scale product.

    This function will be removed from chezmoi after the service is no longer
    available.

[secrets]: https://developer.hashicorp.com/hcp/docs/vault-secrets
[break]: /user-guide/password-managers/hcp-vault-secrets.md#hcp-broken
[eol]: https://support.hashicorp.com/hc/en-us/articles/41802449287955-HCP-Vault-Secrets-End-Of-Life

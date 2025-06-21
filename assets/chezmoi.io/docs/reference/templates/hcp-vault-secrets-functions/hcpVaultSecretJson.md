# `hcpVaultSecretJson` *key* [*application-name* [*project-id* [*organization-id*]]]

`hcpVaultSecretJson` returns structured data from [HCP Vault Secrets][secrets]
using `vlt secrets get --format=json`.

If any of *application-name*, *project-id*, or *organization-id* are empty or
omitted, then chezmoi will use the value from the
`hcpVaultSecret.applicationName`, `hcpVaultSecret.projectId`, and
`hcpVaultSecret.organizationId` config variables if they are set and not empty.

!!! example

    ```
    {{ (hcpVaultSecretJson "secret_name" "application_name").created_by.email }}
    ```

!!! info

    If you access HCP Vault Secrets through the `hcp`, this function **will not
    work**. See [`vlt` vs `hcp`: Upgrades that Break][break].

!!! warning

    Hashicorp has announced the [closure the Vault Secrets service][eol] for
    most customers 27 August 2025 with a recommendation of migrating to
    Hashicorp Vault, which is an enterprise-scale product.

    This function will be removed from chezmoi after the service is no longer
    available.

[secrets]: https://developer.hashicorp.com/hcp/docs/vault-secrets
[break]: /user-guide/password-managers/hcp-vault-secrets.md#hcp-broken
[eol]: https://support.hashicorp.com/hc/en-us/articles/41802449287955-HCP-Vault-Secrets-End-Of-Life

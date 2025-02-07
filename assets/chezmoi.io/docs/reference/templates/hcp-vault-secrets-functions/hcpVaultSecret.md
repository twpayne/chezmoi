# `hcpVaultSecret` *key* [*application-name* [*project-id* [*organization-id*]]]

`hcpVaultSecret` returns the plain text secret from [HCP Vault Secrets][secrets]
using `vlt secrets get --plaintext`.

If any of *application-name*, *project-id*, or *organization-id* are empty or
omitted, then chezmoi will use the value from the
`hcpVaultSecret.applicationName`, `hcpVaultSecret.projectId`, and
`hcpVaultSecret.organizationId` config variables if they are set and not empty.

!!! example

    ```
    {{ hcpVaultSecret "username" }}
    ```

[secrets]: https://developer.hashicorp.com/hcp/docs/vault-secrets

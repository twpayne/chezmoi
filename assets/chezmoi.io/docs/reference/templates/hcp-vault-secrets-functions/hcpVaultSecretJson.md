# `hcpVaultSecretJson` *key* [*application-name* [*project-id* [*organization-id*]]]

`hcpVaultSecretJson` returns structured data from [HCP Vault
Secrets](https://developer.hashicorp.com/hcp/docs/vault-secrets) using `vlt
secrets get --format=json`.

If any of *application-name*, *project-id*, or *organization-id* are empty or omitted, then
chezmoi will use the value from the `hcpVaultSecret.applicationName`,
`hcpVaultSecret.projectId`, and `hcpVaultSecret.organizationId` config variables
if they are set and not empty.

!!! example

    ```
    {{ (hcpVaultSecretJson "secret_name" "application_name").created_by.email }}
    ```

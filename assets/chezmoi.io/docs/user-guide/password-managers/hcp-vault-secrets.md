# HCP Vault Secrets

chezmoi includes support for [HCP Vault
Secrets](https://developer.hashicorp.com/hcp/docs/vault-secrets) using the `vlt`
CLI to expose data through the `hcpVaultSecret` and `hcpVaultSecretJson`
template functions.

Log in using:

```sh
vlt login
```

The output of the `vlt secrets get --plaintext $SECRET_NAME` is available as the
`hcpVaultSecret` function, for example:

```text
{{ hcpVaultSecret "secret_name" "application_name" "project_id" "organization_id" }}
```

You can set the default values for the application name, project ID, and
organization ID in your config file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[hcpVaultSecrets]
    organizationId = "bf479eab-a292-4b46-92df-e22f5c47eadc"
    projectId = "5907a2fa-d26a-462a-8705-74dfe967e87d"
    applicationName = "my-application"
```

With these default values, you can omit them in the call to `hcpVaultSecret`, for example:

```text
{{ hcpVaultSecret "secret_name" }}
{{ hcpVaultSecret "other_secret_name" "other_application_name" }}
```

Structured data from `vlt secrets get --format=json $SECRET_NAME` is available
as the `hcpVaultSecretJson` template function, for example:

```text
{{ (hcpVaultSecretJson "secret_name").created_by.email }}
```

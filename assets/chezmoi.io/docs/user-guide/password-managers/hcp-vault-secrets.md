# HCP Vault Secrets

chezmoi includes support for [HCP Vault Secrets][secrets] using the `vlt` CLI to
expose data through the `hcpVaultSecret` and `hcpVaultSecretJson` template
functions.

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

With these default values, you can omit them in the call to `hcpVaultSecret`,
for example:

```text
{{ hcpVaultSecret "secret_name" }}
{{ hcpVaultSecret "other_secret_name" "other_application_name" }}
```

Structured data from `vlt secrets get --format=json $SECRET_NAME` is available
as the `hcpVaultSecretJson` template function, for example:

```text
{{ (hcpVaultSecretJson "secret_name").created_by.email }}
```

!!! warning

    Hashicorp has announced the [closure the Vault Secrets service][eol] for
    most customers 27 August 2025 with a recommendation of migrating to
    Hashicorp Vault, which is an enterprise-scale product.

    These functions will be removed from chezmoi after the service is no longer
    available.

## `vlt` vs `hcp`: Upgrades that Break { id="hcp-broken" }

Hashicorp ended support for the `vlt` CLI tool in September 2024 and recommended
migrating to the [`hcp` CLI][hcp]. Unfortunately, the new command
does not [work like `vlt`][compat], rendering `hcpVaultSecret` and
`hcpVaultSecretJson` inoperable when using the recommended command-line tool.
[Contributions][contrib] to create new integrations for HCP Vault Secrets are
welcome.

Without these integrations, anyone using HCP Vault Secrets that must upgrade to
the `hcp` client are recommended to use the [`output`][output] and
[`fromJson`][fromjson] functions together:

```ini
{{- $app_name := "my-app-name" -}}
{{- $secret_name := "gdrive-secrets" -}}
{{- $secret :=
    output "hcp" "vs" "s" "open" "--format" "json" "--app" $app_name $secret_name
    | fromJson -}}
[GDrive]
type = drive
client_secret = {{ $secret.static_version.value }}
```

`$HCP_CLIENT_ID` and `$HCP_CLIENT_SECRET` must be set and exported for use in
chezmoi for the above template to work.

See [issue #4146][issue-4146] for more details.

[compat]: https://github.com/twpayne/chezmoi/issues/4146#issuecomment-2552752501
[contrib]: /developer-guide/contributing-changes.md
[fromjson]: /reference/templates/functions/fromJson.md
[hcp]: https://developer.hashicorp.com/hcp/docs/vault-secrets/get-started/install-hcp-cli
[issue-4146]: https://github.com/twpayne/chezmoi/issues/4146
[output]: /reference/templates/functions/output.md
[secrets]: https://developer.hashicorp.com/hcp/docs/vault-secrets
[eol]: https://support.hashicorp.com/hc/en-us/articles/41802449287955-HCP-Vault-Secrets-End-Of-Life

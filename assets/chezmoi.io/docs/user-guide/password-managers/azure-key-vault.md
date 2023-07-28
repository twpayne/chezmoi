# Azure Key Vault

chezmoi includes support for [Azure Key Vault secrets](https://learn.microsoft.com/en-us/azure/key-vault/secrets/about-secrets).

A default Azure Key Vault name can be set in `~/.config/chezmoi/chezmoi.toml` with
`azureKeyVault.defaultVault`.

Ensure [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) is installed and
[log in](https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#azureCLI).
The logged in user must have the `Key Vault Secrets User` RBAC role on the Azure Key Vault resource.

Alternatively, use alternate [authentication options](https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication?tabs=bash#2-authenticate-with-azure).


```toml title="~/.config/chezmoi/chezmoi.toml"
[azureKeyVault]
  defaultVault = "contoso-vault2"
```

A secret value can be retrieved with the `azureKeyVault` template function.

Retrieve the secret `my-secret-name` from the default configured vault.

```
exampleSecret = {{ azureKeyVault "my-secret-name" }}
```

Retrieve the secret `my-secret-name` from the vault named `contoso-vault2`.

```
exampleSecret = {{ azureKeyVault "my-secret-name" "contoso-vault2" }}
```

It is also possible to define an alias in the configuration file for an
additional vault.

```toml title="~/.config/chezmoi/chezmoi.toml"
[data]
  vault42 = "contoso-vault42"

[azureKeyVault]
  defaultVault = "contoso-vault2"
```

Retrieve the secret `my-secret-name` from the vault named `contoso-vault42`
through the alias.

```
exampleSecret = {{ azureKeyVault "my-secret-name" .vault42 }}
```

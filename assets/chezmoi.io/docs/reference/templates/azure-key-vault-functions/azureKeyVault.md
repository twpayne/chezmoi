# `azureKeyVault` *secret name* [*vault-name*]

`azureKeyVault` returns a secret value retrieved from an
[Azure Key Vault](https://learn.microsoft.com/en-us/azure/key-vault/general/).

The mandatory `secret name` argument specifies the *name of the secret* to
retrieve.

The optional `vault name` argument specifies the *name of the vault*, if not set,
the default vault name will be used.

!!! warning

    The current implementation will always return the latest version of the secret.
    Retrieving a specific version of a secret is not supported.

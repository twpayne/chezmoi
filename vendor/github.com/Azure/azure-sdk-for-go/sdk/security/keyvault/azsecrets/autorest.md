## Go

These settings apply only when `--go` is specified on the command line.

```yaml
clear-output-folder: false
export-clients: true
go: true
input-file: https://github.com/Azure/azure-rest-api-specs/blob/551275acb80e1f8b39036b79dfc35a8f63b601a7/specification/keyvault/data-plane/Microsoft.KeyVault/stable/7.4/secrets.json
license-header: MICROSOFT_MIT_NO_VERSION
module: github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets
openapi-type: "data-plane"
output-folder: ../azsecrets
override-client-name: Client
security: "AADToken"
security-scopes: "https://vault.azure.net/.default"
use: "@autorest/go@4.0.0-preview.46"
version: "^3.0.0"
directive:
  # delete unused model
  - remove-model: SecretProperties

  # make vault URL a parameter of the client constructor
  - from: swagger-document
    where: $["x-ms-parameterized-host"]
    transform: $.parameters[0]["x-ms-parameter-location"] = "client"

  # rename parameter models to match their methods
  - rename-model:
      from: SecretRestoreParameters
      to: RestoreSecretParameters
  - rename-model:
      from: SecretSetParameters
      to: SetSecretParameters
  - rename-model:
      from: SecretUpdateParameters
      to: UpdateSecretParameters
  - rename-model:
      from: SecretBundle
      to: Secret
  - rename-model:
      from: DeletedSecretBundle
      to: DeletedSecret
  - rename-model:
      from: SecretItem
      to: SecretProperties
  - rename-model:
      from: DeletedSecretItem
      to: DeletedSecretProperties
  - rename-model:
      from: UpdateSecretParameters
      to: UpdateSecretPropertiesParameters
  - rename-model:
      from: DeletedSecretListResult
      to: DeletedSecretPropertiesListResult
  - rename-model:
      from: SecretListResult
      to: SecretPropertiesListResult

  # rename operations
  - rename-operation:
      from: GetDeletedSecrets
      to: ListDeletedSecretProperties
  - rename-operation:
      from: GetSecrets
      to: ListSecretProperties
  - rename-operation:
      from: GetSecretVersions
      to: ListSecretPropertiesVersions
  - rename-operation:
      from: UpdateSecret
      to: UpdateSecretProperties

  # rename fields
  - from: swagger-document
    where: $.definitions.RestoreSecretParameters.properties.value
    transform: $["x-ms-client-name"] = "SecretBackup"
  - from: swagger-document
    where: $.definitions.Secret.properties.kid
    transform: $["x-ms-client-name"] = "KID"

  # remove type DeletionRecoveryLevel, use string instead
  - from: models.go
    where: $
    transform: return $.replace(/DeletionRecoveryLevel/g, "string");

  # Remove MaxResults parameter
  - where: "$.paths..*"
    remove-parameter:
      in: query
      name: maxresults

  # delete unused error models
  - from: models.go
    where: $
    transform: return $.replace(/(?:\/\/.*\s)+type (?:Error|KeyVaultError).+\{(?:\s.+\s)+\}\s/g, "");
  - from: models_serde.go
    where: $
    transform: return $.replace(/(?:\/\/.*\s)+func \(\w \*?(?:Error|KeyVaultError)\).*\{\s(?:.+\s)+\}\s/g, "");

  # delete the Attributes model defined in common.json (it's used only with allOf)
  - from: models.go
    where: $
    transform: return $.replace(/(?:\/\/.*\s)+type Attributes.+\{(?:\s.+\s)+\}\s/g, "");
  - from: models_serde.go
    where: $
    transform: return $.replace(/(?:\/\/.*\s)+func \(a \*?Attributes\).*\{\s(?:.+\s)+\}\s/g, "");

  # delete the version path param check (version == "" is legal for Key Vault but indescribable by OpenAPI)
  - from: client.go
    where: $
    transform: return $.replace(/\sif secretVersion == "" \{\s+.+secretVersion cannot be empty"\)\s+\}\s/g, "");

  # delete client name prefix from method options and response types
  - from:
      - client.go
      - models.go
      - response_types.go
    where: $
    transform: return $.replace(/Client(\w+)((?:Options|Response))/g, "$1$2");

  # make secret IDs a convenience type so we can add parsing methods
  - from: models.go
    where: $
    transform: return $.replace(/(\sID \*)string(\s+.*)/g, "$1ID$2")
  - from: models.go
    where: $
    transform: return $.replace(/(\sKID \*)string(\s+.*)/g, "$1ID$2")

  # Maxresults -> MaxResults
  - from:
      - client.go
      - models.go
    where: $
    transform: return $.replace(/Maxresults/g, "MaxResults")

  # secretName, secretVersion -> name, version
  - from: client.go
  - where: $
  - transform: return $.replace(/secretName/g, "name").replace(/secretVersion/g, "version")
```

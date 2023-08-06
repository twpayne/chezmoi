# Release History

## 1.0.0 (2023-07-17)

### Features Added
* first stable release of `azsecrets` module

### Breaking Changes
* changed type of `KID` from string to type `ID`

## 0.14.0 (2023-06-08)

### Breaking Changes
* Renamed `Client.ListSecrets` to `Client.ListSecretProperties`
* Renamed `Client.ListSecretVersions` to `Client.ListSecretPropertiesVersions`
* Renamed `SecretBundle` to `Secret`
* Renamed `DeletedSecretBundle` to `DeletedSecret`
* Renamed `SecretItem` to `SecretProperties`
* Renamed `DeletedSecretItem` to `DeletedSecretProperties`
* Renamed `Kid` to `KID`
* Removed `DeletionRecoveryLevel` type
* Remove `MaxResults` option

### Other Changes
* Updated dependencies

## 0.13.0 (2023-04-13)

### Breaking Changes
* Moved from `sdk/keyvault/azsecrets` to `sdk/security/keyvault/azsecrets`

## 0.12.0 (2023-04-13)

### Features Added
* upgraded to api version 7.4

## 0.11.0 (2022-11-08)

### Breaking Changes
* `NewClient` returns an `error`

## 0.10.1 (2022-09-20)

### Features Added
* Added `ClientOptions.DisableChallengeResourceVerification`.
  See https://aka.ms/azsdk/blog/vault-uri for more information.

## 0.10.0 (2022-09-12)

### Breaking Changes
* Verify the challenge resource matches the vault domain.

## 0.9.0 (2022-08-09)

### Breaking Changes
* Changed type of `NewClient` options parameter to `azsecrets.ClientOptions`, which embeds
  the former type, `azcore.ClientOptions`

## 0.8.0 (2022-07-07)

### Breaking Changes
* The `Client` API now corresponds more directly to the Key Vault REST API.
  Most method signatures and types have changed. See the
  [module documentation](https://aka.ms/azsdk/go/keyvault-secrets/docs)
  for updated code examples and more details.

### Other Changes
* Upgrade to latest `azcore`

## 0.7.1 (2022-05-12)

### Other Changes
* Updated to latest `azcore` and `internal` modules.

## 0.7.0 (2022-04-06)

### Features Added
* Added `PossibleDeletionRecoveryLevelValues` to iterate over all valid `DeletionRecoveryLevel` values
* Implemented generic pagers from `runtime.Pager` for all List operations
* Added `Name *string` to `DeletedSecret`, `Properties`, `Secret`, `SecretItem`, and `SecretItem`
* Added `Client.VaultURL` to determine the vault URL for debugging
* Adding `ResumeToken` method to pollers for resuming polling at a later date by using the added `ResumeToken` optional parameter on client polling methods

### Breaking Changes
* Requires a minimum version of go 1.18
* Removed `RawResponse` from pollers
* Removed `DeletionRecoveryLevel`
* Polling operations return a Poller struct directly instead of a Response envelope
* Removed `ToPtr` methods
* `Client.UpdateSecretProperties` takes a `Secret`
* Renamed `Client.ListSecrets` to `Client.ListPropertiesOfSecrets`
* Renamed `Client.ListSecretVersions` to `Client.ListPropertiesOfSecretVersions`
* Renamed `DeletedDate` to `DeletedOn` and `Managed` to `IsManaged`
* Moved `ContentType`, `Tags`, `KeyID`, and `IsManaged` to `Properties`

## 0.6.0 (2022-03-08)

### Breaking Changes
* Changes `Attributes` to `Properties`
* Changes `Secret.KID` to `Secret.KeyID`
* Changes `DeletedSecretBundle` to `DeletedSecret`
* Changes `DeletedDate` to `DeletedOn`, `Created` to `CreatedOn`, and `Updated` to `UpdatedOn`
* Changes the signature of `Client.UpdateSecretProperties` to have all alterable properties in the `UpdateSecretPropertiesOptions` parameter, removing the `parameters Properties` parameter.
* Changes `Item` to `SecretItem`
* Pollers and pagers are structs instead of interfaces
* Prefixed all `DeletionRecoveryLevel` constants with "DeletionRecoveryLevel"
* Changed pager APIs for `ListSecretVersionsPager`, `ListDeletedSecretsPager`, and `ListSecretsPager`
    * Use the `More()` method to determine if there are more pages to fetch
    * Use the `NextPage(context.Context)` to fetch the next page of results
* Removed all `RawResponse *http.Response` fields from response structs.

## 0.5.0 (2022-02-08)

### Breaking Changes
* Fixes a bug where `UpdateSecretProperties` will delete properties that are not explicitly set each time. This is only a breaking change at runtime, where the request body will change.

## 0.4.0 (2022-01-11)

### Other Changes
* Bumps `azcore` dependency from `v0.20.0` to `v0.21.0`

## 0.3.0 (2021-11-09)

### Features Added
* Clients can now connect to Key Vaults in any cloud

## 0.2.0 (2021-11-02)

### Other Changes
* Bumps `azcore` dependency to `v0.20.0` and `azidentity` to `v0.12.0`

## 0.1.1 (2021-10-06)
* Adds the MIT License for redistribution

## 0.1.0 (2021-10-05)
* This is the initial release of the `azsecrets` library

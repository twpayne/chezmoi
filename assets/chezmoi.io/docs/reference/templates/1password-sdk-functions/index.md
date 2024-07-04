# 1Password SDK functions

!!! warning

    1Password SDK template functions are experimental and may change.

The `onepasswordSDK*` template functions return structured data from
[1Password](https://1password.com/) using the [1Password
SDK](https://developer.1password.com/docs/sdks/).

By default, the 1Password service account token is taken from the
`$OP_SERVICE_ACCOUNT_TOKEN` environment variable. The name of the environment
variable can be set with `onepasswordSDK.tokenEnvVar` configuration variable, or
the token can be set explicitly by setting the `onepasswordSDK.token`
configuration variable.

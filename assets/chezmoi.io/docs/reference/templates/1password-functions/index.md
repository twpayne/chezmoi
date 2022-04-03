# 1Password functions

The `onepassword*` template functions return structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`).

!!! warning

    When using 1Password CLI 2.0, there may be an issue with pre-authenticating
    `op` because the environment variable used to store the session key has
    changed from `OP_SESSION_account` to `OP_SESSION_accountUUID`. Instead of
    using *account-name*, it is recommended that you use the *account-uuid*.
    This can be found using `op account list`.

    This issue does not occur when using biometric authentication and 1Password
    8, or if you allow chezmoi to prompt you for 1Password authentication
    (`1password.prompt = true`).

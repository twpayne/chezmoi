# 1Password functions

The `onepassword*` template functions return structured data from
[1Password](https://1password.com/) using the [1Password
CLI](https://developer.1password.com/docs/cli) (`op`).

!!! info

    When using the 1Password CLI with biometric authentication, chezmoi derives
    values from `op account list` that can resolves into the appropriate
    1Password *account-uuid*.

    As an example, if `op account list --format=json` returns the following
    structure:

    ```json
    [
      {
        "url": "account1.1password.ca",
        "email": "my@email.com",
        "user_uuid": "some-user-uuid",
        "account_uuid": "some-account-uuid"
      }
    ]
    ```

    The following values can be used in the `account` parameter and the value
    `some-account-uuid` will be passed as the `--account` parameter to `op`.

    - `some-account-uuid`
    - `some-user-uuid`
    - `account1.1password.ca`
    - `account1`
    - `my@email.com`
    - `my`
    - `my@account1.1password.ca`
    - `my@account1`

    If there are multiple accounts and _any_ value exists more than once, that
    value will be removed from the account mapping. That is, if you are signed
    into `my@email.com` and `your@email.com` for `account1.1password.ca`, then
    `account1.1password.ca` will not be a valid lookup value, but `my@account1`,
    `my@account1.1password.ca`, `your@account1`, and
    `your@account1.1password.ca` would all be valid lookups.

!!! warning

    chezmoi has experimental support for [1Password secrets
    automation](../../../user-guide/password-managers/1password.md#secrets-automation)
    modes. These modes change how the 1Password CLI works and affect all
    functions. Most notably, `account` parameters are not allowed on all
    1Password template functions.

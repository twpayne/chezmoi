# `onepassword` *uuid* [*vault-uuid* [*account-name*]]

`onepassword` returns structured data from [1Password](https://1password.com/)
using the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`).
*uuid* is passed to `op get item $UUID` and the output from `op` is parsed as
JSON. The output from `op` is cached so calling `onepassword` multiple times
with the same *uuid* will only invoke `op` once.  If the optional *vault-uuid*
is supplied, it will be passed along to the `op get` call, which can
significantly improve performance. If the optional *account-name* is supplied,
it will be passed along to the `op get` call, which will help it look in the
right account, in case you have multiple accounts (eg. personal and work
accounts). If there is no valid session in the environment, by default you will
be interactively prompted to sign in.

!!! example

    ```
    {{ (onepassword "$UUID").details.password }}
    {{ (onepassword "$UUID" "$VAULT_UUID").details.password }}
    {{ (onepassword "$UUID" "$VAULT_UUID" "$ACCOUNT_NAME").details.password }}
    ```

    If using 1Password 1.0, then *vault-uuid* is optional.

    ```
    {{ (onepassword "$UUID" "" "$ACCOUNT_NAME").details.password }}
    ```

!!! info

    If you're using [1Password CLI 2.0](https://developer.1password.com/), there
    are changes to be aware of.

    !!! warning

    The structure of the data returned by the `onepassword` template function
    will be different and you will need to update your templates. The structure
    has not yet been finalized.

    !!! warning

    Neither *vault-uuid* nor *account-name* may be empty strings if specified.
    Older versions of 1Password CLI would ignore empty strings for arguments.

    !!! warning

    Unless using biometric authentication, or when using without prompting, it
    is recommended that instead of *account-name*, the UUID of the account is
    used. This can be shown with `op account list`.

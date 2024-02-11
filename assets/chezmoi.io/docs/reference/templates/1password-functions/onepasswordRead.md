# `onepasswordRead` *url* [*account*]

`onepasswordRead` returns data from [1Password](https://1password.com/) using
the [1Password CLI](https://developer.1password.com/docs/cli) (`op`). *url* is
passed to the `op read --no-newline` command. If *account* is specified, the
extra arguments `--account $ACCOUNT` are passed to `op`.

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

!!! example

    The result of

    ```
    {{ onepasswordRead "op://vault/item/field" }}
    ```

    is equivalent to calling

    ```console
    $ op read --no-newline op://vault/item/field
    ```

!!! warning

    When using [1Password secrets
    automation](../../user-guide/password-managers/1password.md#secrets-automation),
    the *account* parameter is not allowed.

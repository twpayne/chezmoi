# `onepasswordRead` *url* [*account*]

`onepasswordRead` returns data from [1Password][1p] using the [1Password
CLI][op] (`op`). *url* is passed to the `op read --no-newline` command. If
*account* is specified, the extra arguments `--account $ACCOUNT` are passed to
`op`.

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

    When using [1Password secrets automation][automation], the *account*
    parameter is not allowed.

[1p]: https://1password.com/
[op]: https://developer.1password.com/docs/cli
[automation]: /user-guide/password-managers/1password.md#secrets-automation

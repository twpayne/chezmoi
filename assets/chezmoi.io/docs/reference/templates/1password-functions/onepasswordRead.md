# `onepasswordRead` *url* [*account*]

`onepasswordRead` returns data from [1Password](https://1password.com/) using
the [1Password
CLI](https://support.1password.com/command-line-getting-started/) (`op`). *url*
is passed to `op read $URL`. If *account* is specified, the extra arguments
`--account $ACCOUNT` are passed to `op`.

If there is no valid session in the environment, by default you will be
interactively prompted to sign in.

!!! example

    The result of

    ```
    {{ onepasswordRead "op://vault/item/field" }}
    ```

    is equivalent to calling

    ```console
    $ op read op://vault/item/field
    ```

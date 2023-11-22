# `forget` *target*...

Remove *target*s from the source state, i.e. stop managing them. *target*s must
have entries in the source state. They cannot be externals.

!!! example

    ```console
    $ chezmoi forget ~/.bashrc
    ```

!!! info

    To remove targets from both the source state and destination directory, use
    [`remove`](/reference/commands/remov).

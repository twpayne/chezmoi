# `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with a custom secret
manager. Normally you would use chezmoi's existing template functions to retrieve secrets.

!!! note

    If you need to pass flags to the secret manager's CLI you must separate
    them with `--` to prevent chezmoi from interpreting them.

!!! hint

    To get a full list of subcommands run:

    ```console
    $ chezmoi secret help
    ```

!!! example

    ```console
    $ chezmoi secret keyring set --service=service --user=user --value=password
    $ chezmoi secret keyring get --service=service --user=user
    $ chezmoi secret keyring delete --service=service --user=user
    ```

!!! warning

    On FreeBSD, the `secret keyring` command is only available if chezmoi was
    compiled with cgo enabled. The official release binaries of chezmoi are
    **not** compiled with cgo enabled, and `secret keyring` command is not
    available.

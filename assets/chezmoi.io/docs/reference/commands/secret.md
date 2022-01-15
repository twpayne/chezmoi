# `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with your secret
manager. Normally you would use template functions to retrieve secrets. Note
that if you want to pass flags to the secret manager's CLI you will need to
separate them with `--` to prevent chezmoi from interpreting them.

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

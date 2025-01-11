# `secret`

Run a secret manager's CLI, passing any extra arguments to the secret manager's
CLI. This is primarily for verifying chezmoi's integration with a custom secret
manager. Normally you would use chezmoi's existing template functions to retrieve secrets.

!!! note

    If you need to pass flags to the secret manager's CLI you must separate
    them with `--` to prevent chezmoi from interpreting them.

## Subcommands

### `secret keyring delete`

#### `--service` *string*

Name of the service.

#### `--user` *string*

Name of the user.

### `secret keyring get`

#### `--service` *string*

Name of the service.

#### `--user` *string*

Name of the user.

### `secret keyring set`

#### `--service` *string*

Name of the service.

#### `--user` *string*

Name of the user.

#### `--value` *string*

New value.

## Examples

```sh
chezmoi secret keyring set --service=service --user=user --value=password
chezmoi secret keyring get --service=service --user=user
chezmoi secret keyring delete --service=service --user=user
```

## Notes

!!! warning

    On FreeBSD, the `secret keyring` command is only available if chezmoi was
    compiled with cgo enabled. The official release binaries of chezmoi are
    **not** compiled with cgo enabled, and `secret keyring` command is not
    available.

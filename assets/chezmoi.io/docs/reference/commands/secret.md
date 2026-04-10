# `secret`

<!-- markdownlint-disable no-duplicate-heading -->

Verify chezmoi's integration with the [system's keyring][keyring].

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

[keyring]: /user-guide/password-managers/keychain-and-windows-credentials-manager.md

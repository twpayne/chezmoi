# `state`

Manipulate the persistent state.

!!! hint

    To get a full list of subcommands run:

    ```console
    $ chezmoi state help
    ```

## Subcommands

### `data`

Print the raw data in the persistent state.

### `delete`

Delete a value from the persistent state.

### `delete-bucket`

Delete a bucket from the persistent state.

### `dump`

Generate a dump of the persistent state.

### `get`

Get a value from the persistent state.

### `get-bucket`

Get a bucket from the persistent state.

### `reset`

Reset the persistent state.

### `set`

Set a value from the persistent state

## Examples

```sh
chezmoi state data
chezmoi state delete --bucket=$BUCKET --key=$KEY
chezmoi state delete-bucket --bucket=$BUCKET
chezmoi state dump
chezmoi state get --bucket=$BUCKET --key=$KEY
chezmoi state get-bucket --bucket=$BUCKET
chezmoi state set --bucket=$BUCKET --key=$KEY --value=$VALUE
chezmoi state reset
```

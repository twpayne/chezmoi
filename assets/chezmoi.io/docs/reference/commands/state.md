# `state`

Manipulate the persistent state.

!!! hint

    To get a full list of subcommands run:

    ```console
    $ chezmoi state help
    ```

# Available subcommands

## `data`

Print the raw data in the persistent state.

## `delete`

Delete a value from the persistent state.

## `delete-bucket`

Delete a bucket from the persistent state.

## `dump`

Generate a dump of the persistent state.

## `get`

Get a value from the persistent state.

## `get-bucket`

Get a bucket from the persistent state.

## `reset`

Reset the persistent state.

## `set`

Set a value from the persistent state

!!! example

    ```console
    $ chezmoi state data
    $ chezmoi state delete --bucket=bucket --key=key
    $ chezmoi state delete-bucket --bucket=bucket
    $ chezmoi state dump
    $ chezmoi state get --bucket=bucket --key=key
    $ chezmoi state get-bucket --bucket=bucket
    $ chezmoi state set --bucket=bucket --key=key --value=value
    $ chezmoi state reset
    ```

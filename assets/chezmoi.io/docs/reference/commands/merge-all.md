# `merge-all`

Perform a three-way merge for file whose actual state does not match its target
state. The merge is performed with `chezmoi merge`.

A custom merge command may be configured by specifying an optional `merge`
section within the `mergeAll` configuration section. See `merge` for additional
details.

!!! example

    ```console
    $ chezmoi merge-all
    ```

!!! example

    ```toml
    [mergeAll]
        [merge]
            Command = my-custom-merge-all-tool
            Args = [ {{ .Destination }}, {{ .Source }}, {{ .Target }}]
    ```

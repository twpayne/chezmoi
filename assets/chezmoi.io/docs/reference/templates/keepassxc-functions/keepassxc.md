# `keepassxc` *entry*

`keepassxc` returns structured data for *entry* using `keepassxc-cli`.

The output from `keepassxc-cli` is parsed into key-value pairs and cached so
calling `keepassxc` multiple times with the same *entry* will only invoke
`keepassxc-cli` once.

!!! example

    ```
    username = {{ (keepassxc "example.com").UserName }}
    password = {{ (keepassxc "example.com").Password }}
    ```

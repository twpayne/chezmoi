# `keepassxc` *entry*

`keepassxc` returns structured data retrieved from a
[KeePassXC](https://keepassxc.org/) database using the KeePassXC CLI
(`keepassxc-cli`). The database is configured by setting `keepassxc.database`
in the configuration file. *database* and *entry* are passed to `keepassxc-cli
show`. You will be prompted for the database password the first time
`keepassxc-cli` is run, and the password is cached, in plain text, in memory
until chezmoi terminates. The output from `keepassxc-cli` is parsed into
key-value pairs and cached so calling `keepassxc` multiple times with the same
*entry* will only invoke `keepassxc-cli` once.

!!! example

    ```
    username = {{ (keepassxc "example.com").UserName }}
    password = {{ (keepassxc "example.com").Password }}
    ```

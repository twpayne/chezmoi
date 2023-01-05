# KeePassXC functions

The `keepassxc*` template functions return structured data retrieved from a
[KeePassXC](https://keepassxc.org/) database using the KeePassXC CLI
(`keepassxc-cli`)

The database is configured by setting `keepassxc.database` in the configuration
file. You will be prompted for the database password the first time
`keepassxc-cli` is run, and the password is cached, in plain text, in memory
until chezmoi terminates.

The command used can by changed by setting the `keepassxc.command`
configuration variable, and extra arguments can be added by setting
`keepassxc.args`. Also, you can disable the password prompt by setting
`keepassxc.prompt` to `false`.

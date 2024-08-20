# KeePassXC functions

The `keepassxc*` template functions return structured data retrieved from a
[KeePassXC](https://keepassxc.org/) database using the KeePassXC CLI
(`keepassxc-cli`)

The database is configured by setting `keepassxc.database` in the configuration
file. You will be prompted for the database password the first time
`keepassxc-cli` is run, and the password is cached, in plain text, in memory
until chezmoi terminates.

The command used can by changed by setting the `keepassxc.command` configuration
variable, and extra arguments can be added by setting `keepassxc.args`. The
password prompt can be disabled by setting `keepassxc.prompt` to `false`.

By default, chezmoi will prompt for the KeePassXC password when required and
cache it for the duration of chezmoi's execution. Setting `keepassxc.mode` to
`open` will tell chezmoi to instead open KeePassXC's console with `keepassxc-cli
open` followed by `keepassxc.args`. chezmoi will use this console to request
values from KeePassXC.

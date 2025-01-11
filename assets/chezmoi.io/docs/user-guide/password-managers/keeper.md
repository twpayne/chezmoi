# Keeper

chezmoi includes support for [Keeper](https://www.keepersecurity.com/) using the
[Commander CLI](https://docs.keeper.io/secrets-manager/commander-cli) to expose
data as a template function.

Create a persistent login session as [described in the Command CLI
documentation](https://docs.keeper.io/secrets-manager/commander-cli/using-commander/logging-in#persistent-login-sessions).

Passwords can be retrieved with the `keeperFindPassword` template function, for
example:

```text
examplePasswordFromPath = {{ keeperFindPassword "$PATH" }}
examplePasswordFromUid = {{ keeperFindPassword "$UID" }}
```

For retrieving more complex data, use the `keeper` template function with a UID
to retrieve structured data from [`keeper
get`](https://docs.keeper.io/secrets-manager/commander-cli/using-commander/command-reference/record-commands#get-command)
or the `keeperDataFields` template function which restructures the output of
`keeper get` in to a more convenient form, for example:

```text
keeperDataTitle = {{ (keeper "$UID").data.title }}
examplePassword = {{ index (keeperDataFields "$UID").password 0 }}
```

Extra arguments can be passed to the Keeper CLI command by setting the
`keeper.args` variable in chezmoi's config file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[keeper]
    args = ["--config", "/path/to/config.json"]
```

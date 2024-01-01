# KeePassXC

chezmoi includes support for [KeePassXC](https://keepassxc.org) using the
KeePassXC CLI (`keepassxc-cli`) to expose data as a template function.

Provide the path to your KeePassXC database in your configuration file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[keepassxc]
    database = "/home/user/Passwords.kdbx"
```

The structured data from `keepassxc-cli show $database` is available as the
`keepassxc` template function in your config files, for example:

```
username = {{ (keepassxc "example.com").UserName }}
password = {{ (keepassxc "example.com").Password }}
```

Additional attributes are available through the `keepassxcAttribute` function.
For example, if you have an entry called `SSH Key` with an additional attribute
called `private-key`, its value is available as:

```
{{ keepassxcAttribute "SSH Key" "private-key" }}
```

## Non-password-protected databases

If your database is not password protected, add `--no-password` to
`keepassxc.args` and `keepassxc.prompt = false`:

```toml title="~/.config/chezmoi/chezmoi.toml"
[keepassxc]
    args = ["--no-password"]
    prompt = false
```

## YubiKey support

chezmoi includes an experimental mode to support using KeePassXC with YubiKeys.
Set `keepassxc.mode` to `open` and `keepassxc.args` to the arguments required to
set your YubiKey, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[keepassxc]
    args = ["--yubikey", "1:7370001"]
    mode = "open"
```

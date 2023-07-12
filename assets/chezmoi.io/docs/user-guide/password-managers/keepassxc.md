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

!!! info

    The KeePassXC CLI does not currently support any persistent login, which
    means that you will have to enter your password every time you run chezmoi.

    Support for non-password authentication methods (e.g. YubiKey) is currently
    poor. For more information, see [this GitHub
    issue](https://github.com/twpayne/chezmoi/issues/2002).

# age

chezmoi supports encrypting files with [age](https://age-encryption.org/).

Generate a key using `age-keygen`:

```console
$ age-keygen -o $HOME/key.txt
Public key: age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

Specify age encryption in your configuration file, being sure to specify at
least the identity and one recipient:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    identity = "/home/user/key.txt"
    recipient = "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"
```

chezmoi supports multiple recipients and recipient files, and multiple
identities.

## Symmetric encryption

To use age's symmetric encryption, specify a single identity and enable
symmetric encryption in your config file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    identity = "~/.ssh/id_rsa"
    symmetric = true
```

## Symmetric encryption with a passphrase

To use age's symmetric encryption with a passphrase, set `age.passphrase` to
`true` in your config file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    passphrase = true
```

You will be prompted for the passphrase whenever you run `chezmoi add
--encrypt` and whenever chezmoi needs to decrypt the file, for example when you
run `chezmoi apply`, `chezmoi diff`, or `chezmoi status`.

## Builtin age encryption

chezmoi has builtin support for age encryption which is automatically used if
the `age` command is not found in `$PATH`.

!!! info

    The builtin age encryption does not support SSH keys.

    SSH keys are not supported as the [age documentation explicitly recommends
    not using them](https://pkg.go.dev/filippo.io/age#hdr-Key_management):

    > When integrating age into a new system, it's recommended that you only
    > support X25519 keys, and not SSH keys. The latter are supported for
    > manual encryption operations.

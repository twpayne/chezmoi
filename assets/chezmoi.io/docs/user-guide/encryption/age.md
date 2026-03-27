# age

chezmoi supports encrypting files with [age][age].

Generate a key using `chezmoi age-keygen`:

```console
$ chezmoi age-keygen --output=$HOME/key.txt
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

chezmoi supports multiple identities and multiple recipients:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    identities = ["/home/user/key1.txt", "/home/user/key2.txt"]
    recipients = ["recipient1", "recipient2"]
```

!!! note

    Make sure `encryption` is added to the top level section at the beginning of
    the config, before any other sections.

## Using identities without a recipient

If you have an identity file without a recipient, e.g. when you generate a Yubikey FIDO2 backed age identity with `age-plugin-fido2prf` like

```shell
age-plugin-fido2prf -generate test > fido.txt`
```

you can set

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    identities = ["identity.txt", "fido.txt"]
    recipients = ["recipient1"]
    useIdentitiesForEncryption = true
```

to use all identities which exist on disk when **encrypting**.

!!! note

    All identities are used when decrypting.

## Symmetric encryption

To use age's symmetric encryption, you do the same as described in the above section
[above section](#using-identities-with-not-recipient).

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    identity = "~/.ssh/id_rsa"
    useIdentitiesForEncryption = true
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

    The builtin age encryption does not support `passphrase = true`,
    `useIdentitiesForEncryption = true` or SSH keys.

    Passphrases are not supported because chezmoi needs to decrypt files
    regularly, e.g. when running a `chezmoi diff` or a `chezmoi status`
    command, not just when running `chezmoi apply`. Prompting for a passphrase
    each time would quickly become tiresome.

    Symmetric encryption may be supported in the future. Please [open an
    issue][issue] if you want this.

    SSH keys are not supported as the [age documentation explicitly recommends
    not using them][nossh]:

    > When integrating age into a new system, it's recommended that you only
    > support X25519 keys, and not SSH keys. The latter are supported for
    > manual encryption operations.

[age]: https://age-encryption.org/
[issue]: https://github.com/twpayne/chezmoi/issues/new?assignees=&labels=enhancement&template=02_feature_request.md&title=
[nossh]: https://pkg.go.dev/filippo.io/age#hdr-Key_management

# gpg

chezmoi supports encrypting files with [gpg](https://www.gnupg.org/). Encrypted
files are stored in the source state and automatically be decrypted when
generating the target state or editing a file contents with `chezmoi edit`.

## Asymmetric (private/public-key) encryption

Specify the encryption key to use in your configuration file (`chezmoi.toml`)
with the `gpg.recipient` key:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "gpg"
[gpg]
    recipient = "..."
```

chezmoi will encrypt files:

```sh
gpg --armor --recipient $RECIPIENT --encrypt
```

and store the encrypted file in the source state. The file will automatically
be decrypted when generating the target state.

## Symmetric encryption

Specify symmetric encryption in your configuration file:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "gpg"
[gpg]
    symmetric = true
```

chezmoi will encrypt files:

```sh
gpg --armor --symmetric
```

## Encrypting files with a passphrase

If you want to encrypt your files with a passphrase, but don't mind the
passphrase being stored in plaintext on your machines, then you can use the
following configuration:

``` title="~/.local/share/chezmoi/.chezmoi.toml.tmpl"
{{ $passphrase := promptStringOnce . "passphrase" "passphrase" -}}

encryption = "gpg"
[data]
    passphrase = {{ $passphrase | quote }}
[gpg]
    symmetric = true
    args = ["--batch", "--passphrase", {{ $passphrase | quote }}, "--no-symkey-cache"]
```

This will prompt you for the passphrase the first time you run `chezmoi init` on
a new machine, and then remember the passphrase in your configuration file.

## Muting gpg output

Since gpg sends some info messages to stderr instead of stdout, you will see
some output even if you redirect stdout to `/dev/null`.

You can mute this by adding `--quiet` to the `gpg.args` key in your
configuration:

```toml title="~/.local/share/chezmoi/.chezmoi.toml.tmpl"
[gpg]
    args = ["--quiet"]
```

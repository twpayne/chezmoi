# gpg

chezmoi supports encrypting files with [gpg](https://www.gnupg.org/). Encrypted
files are stored in the source state and automatically be decrypted when
generating the target state or printing a file's contents with `chezmoi cat`.

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
gpg --armor --recipient <recipient> --encrypt
```

and store the encrypted file in the source state. The file will automatically
be decrypted when generating the target state.

!!! hint

    The `gpg.recipient` key must be ultimately trusted, otherwise encryption
    will fail because gpg will prompt for input, which chezmoi does not handle.
    You can check the trust level by running:

    ```console
    $ gpg --export-ownertrust
    ```

    The trust level for the recipient's key should be `6`. If it is not, you
    can change the trust level by running:

    ```console
    $ gpg --edit-key <recipient>
    ```

    Enter `trust` at the prompt and chose `5 = I trust ultimately`.

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

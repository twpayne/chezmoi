# `age`

<!-- markdownlint-disable no-duplicate-heading -->

Interact with age's passphrase-based encryption.

## Subcommands

### `age encrypt` [*file*...]

Encrypt file or standard input.

#### `-p`, `--passphrase`

Decrypt with a passphrase.

### `age decrypt` [*file*...]

Decrypt file or standard input.

#### `-p`, `--passphrase`

Decrypt with a passphrase.

## Examples

```sh
chezmoi age encrypt --passphrase plaintext.txt > ciphertext.txt
chezmoi age decrypt --passphrase ciphertext.txt > decrypted-ciphertext.txt
```

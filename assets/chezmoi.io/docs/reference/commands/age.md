# `age`

Interact with age's passphrase-based encryption.

# `age encrypt` [*file*...]

## `-p`, `--passphrase`

Encrypt with a passphrase.

# `age decrypt` [*file*...]

## `-p`, `--passphrase`

Decrypt with a passphrase.

!!! example

    ```console
    $ chezmoi age encrypt --passphrase plaintext.txt > ciphertext.txt
    $ chezmoi age decrypt --passphrase ciphertext.txt > decrypted-ciphertext.txt
    ```

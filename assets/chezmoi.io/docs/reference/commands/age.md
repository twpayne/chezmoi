# `age`

Interact with age's passphrase-based encryption.

!!! hint

    To get a full list of subcommands run:

    ```console
    $ chezmoi age help
    ```

!!! example

    ```console
    $ chezmoi age encrypt --passphrase plaintext.txt > ciphertext.txt
    $ chezmoi age decrypt --passphrase ciphertext.txt > decrypted-ciphertext.txt
    ```

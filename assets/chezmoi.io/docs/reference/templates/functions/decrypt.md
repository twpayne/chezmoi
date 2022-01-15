# `decrypt` *ciphertext*

`decrypt` decrypts *ciphertext* using chezmoi's configured encryption method.

!!! example

    ```
    {{ joinPath .chezmoi.sourceDir ".ignored-encrypted-file.age" | include | decrypt }}
    ```

# `ejsonDecrypt` *filePath*

`ejsonDecrypt` returns the decrypted content of an
[ejson](https://github.com/Shopify/ejson)-encrypted file.

*filePath* indicates where the encrypted file is located.

The decrypted file is cached so calling `ejsonDecrypt` multiple
times with the same *filePath* will only run through the decryption
process once. The cache is shared with `ejsonDecryptWithKey`.

!!! example

    ```
    {{ (ejsonDecrypt "my-secrets.ejson").password }}
    ```

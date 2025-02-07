# `ejsonDecryptWithKey` *filePath* *key*

`ejsonDecryptWithKey` returns the decrypted content of an
[ejson][ejson]-encrypted file.

*filePath* indicates where the encrypted file is located,
and *key* is used to decrypt the file.

The decrypted file is cached so calling `ejsonDecryptWithKey` multiple
times with the same *filePath* will only run through the decryption
process once. The cache is shared with `ejsonDecrypt`.

!!! example

    ```
    {{ (ejsonDecryptWithKey "my-secrets.ejson" "top-secret-key").password }}
    ```

[ejson]: https://github.com/Shopify/ejson

# Encryption

chezmoi supports encrypting files with [age](https://age-encryption.org)
and [gpg](https://www.gnupg.com/).

Encrypted files are stored in ASCII-armored format in the source directory with
the `encrypted_` attribute and are automatically decrypted when needed.

Add files to be encrypted with the `--encrypt` flag, for example:

```sh
chezmoi add --encrypt ~/.ssh/id_rsa
```

`chezmoi edit` will transparently decrypt the file before editing and
re-encrypt it afterwards.

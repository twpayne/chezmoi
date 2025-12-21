# Encryption

chezmoi supports encrypting files with [age][age], [git-crypt][gitcrypt],
[gpg][gpg], and [transcrypt][transcrypt].

Encrypted files are stored in ASCII-armored format in the source directory with
the `encrypted_` attribute and are automatically decrypted when needed.

Add files to be encrypted with the `--encrypt` flag, for example:

```sh
chezmoi add --encrypt ~/.ssh/id_rsa
```

`chezmoi edit` will transparently decrypt the file before editing and
re-encrypt it afterwards.

[age]: https://age-encryption.org
[gitcrypt]: https://github.com/AGWA/git-crypt
[gpg]: https://www.gnupg.com/
[transcrypt]: https://github.com/elasticdog/transcrypt

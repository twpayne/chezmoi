# `edit-encrypted` *filename*...

Edit the encrypted files *filename*s.

Each *filename* is decrypted to a temporary directory, the editor is invoked on
the decrypted files. After the editor returns, each the decrypted file is
re-encrypted.

## Examples

  ```sh
  chezmoi edit-encrypted encrypted_file
  ```

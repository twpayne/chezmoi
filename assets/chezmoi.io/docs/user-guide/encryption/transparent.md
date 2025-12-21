# Transparent

chezmoi supports encrypting files with transparent git encryption tools like
[transcrypt][transcrypt] and [git-crypt][gitcrypt].

## transcrypt

In your configuration file, set `encryption` to `transparent`:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "transparent"
```

Initialize transcrypt:

```console
$ chezmoi cd
$ transcrypt
```

Edit `.gitattributes` to use transcrypt for files with the `encrypted_` prefix:

```gitattributes title="~/.local/share/chezmoi/.gitattributes"
encrypted_* filter=crypt diff=crypt merge=crypt
```

Add an encrypted file to both chezmoi and git:

```console
$ chezmoi add ~/.config/sensitive_file
$ git add dot_config/encrypted_sensitive_file
$ git commit -m "Add .config/sensitive_file"
```

Verify that the file is handled by transcrypt:

```console
$ git ls-crypt
dot_config/encrypted_sensitive_file
```

Note that commands like `git show`, `git diff`, etc. will also show the
cleartext form of the file.

Use `transcrypt --display` to show instructions for how to setup transcrypt
after cloning the repository elsewhere. It will involve running a command like:

```
$ transcrypt -c aes-256-cbc -p $PASSWORD
```

[gitcrypt]: https://github.com/AGWA/git-crypt
[transcrypt]: https://github.com/elasticdog/transcrypt

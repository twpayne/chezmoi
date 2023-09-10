# Encryption

## How do I configure chezmoi to encrypt files but only request a passphrase the first time `chezmoi init` is run?

The following steps use [age](https://age-encryption.org/) for encryption.

This can be achieved with the following process:

1. Generate an age private key.
2. Encrypt the private key with a passphrase.
3. Configure chezmoi to decrypt the private key if needed.
4. Configure chezmoi to use the private key.
5. Add encrypted files.

First, change to chezmoi's root directory:

```console
$ chezmoi cd ~
```

Generate an age private key encrypted with a passphrase in the file
`key.txt.age` with the command:

```console
$ age-keygen | age --passphrase > key.txt.age
Public key: age193wd0hfuhtjfsunlq3c83s8m93pde442dkcn7lmj3lspeekm9g7stwutrl
Enter passphrase (leave empty to autogenerate a secure one):
Confirm passphrase:
```

Use a strong passphrase and make a note of the public key
(`age193wd0hfuhtjfsunlq3c83s8m93pde442dkcn7lmj3lspeekm9g7stwutrl` in this case).

Add `key.txt.age` to `.chezmoiignore` so that chezmoi does not try to create it:

```console
$ echo key.txt.age >> .chezmoiignore
```

Configure chezmoi to decrypt the passphrase-encrypted private key if needed:

```console
$ cat > run_once_before_decrypt-private-key.sh.tmpl <<EOF
#!/bin/sh

if [ ! -f "${HOME}/.config/chezmoi/key.txt" ]; then
    mkdir -p "${HOME}/.config/chezmoi"
    chezmoi age decrypt --output "${HOME}/.config/chezmoi/key.txt" --passphrase "{{ .chezmoi.sourceDir }}/key.txt.age"
    chmod 600 "${HOME}/key.txt"
fi
EOF
```

Configure chezmoi to use the public and private key for encryption:

```console
$ cat >> .chezmoi.toml.tmpl <<EOF
encryption = "age"
[age]
    identity = "~/.config/chezmoi/key.txt"
    recipient = "age193wd0hfuhtjfsunlq3c83s8m93pde442dkcn7lmj3lspeekm9g7stwutrl"
EOF
```

`age.recipient` must be your public key from above.

Run `chezmoi init --apply` to generate the chezmoi's config file and decrypt the
private key:

```console
$ chezmoi init --apply
Enter passphrase:
```

At this stage everything is configured and `git status` should report:

```console
$ git status
On branch main
Untracked files:
  (use "git add <file>..." to include in what will be committed)
	.chezmoi.toml.tmpl
	.chezmoiignore
	key.txt.age
	run_once_before_decrypt-private-key.sh.tmpl

nothing added to commit but untracked files present (use "git add" to track)
```

If you're happy with the changes you can commit them. All four files should be
committed.

Add files that you want to encrypt using the `--encrypt` argument to `chezmoi
add`, for example:

```console
$ chezmoi add --encrypt ~/.ssh/id_rsa
```

When you run `chezmoi init` on a new machine you will be prompted to enter your
passphrase once to decrypt `key.txt.age`. Your decrypted private key will be
stored in `~/.config/chezmoi/key.txt`.

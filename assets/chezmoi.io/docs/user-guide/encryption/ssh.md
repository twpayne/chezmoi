# SSH

chezmoi supports encrypting files using SSH keys via `ssh-agent`. This allows you to use your existing SSH identity for encryption without needing GPG or Age keys.

## Configuration

Specify `ssh` encryption in your configuration file, and provide the public key identity you wish to use:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "ssh"
[ssh]
    identity = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
```

You can obtain your public key string by running:

```console
$ ssh-add -L
```

## How it works

chezmoi uses a "Challenge-Response" mechanism to derive a secure encryption key from your SSH key:

1.  **Encryption**: chezmoi generates a random challenge. Your `ssh-agent` signs this challenge. The signature is used to derive the encryption key (using scrypt).
2.  **Decryption**: chezmoi asks the agent to sign the challenge again. If the correct key is loaded, the signature matches, and the key is derived.

## Portability

*   **Agent Forwarding**: You can decrypt files on a remote server without copying your private key by using SSH Agent Forwarding (`ssh -A`).
*   **Hardware Keys**: Works seamlessly with hardware tokens (YubiKey, etc.) that are exposed via `ssh-agent`.

!!! warning

    This method requires the **exact same private key** to be present in the agent for decryption. It does not support encrypting for multiple recipients/keys simultaneously in this version.

# `age-keygen` [*identity-file*]

Generate an age identity or convert an age identity to an age recipient.

## Flags

### `-y`, `--convert`

Read an identity file *identity-file* or the standard input and print its
recipient instead of generating an age identity.

## Examples

```sh
chezmoi age-keygen
chezmoi age-keygen -o identity.txt
chezmoi age-keygen -y identity.txt
```

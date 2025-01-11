# `forget` *target*...

Remove *target*s from the source state, i.e. stop managing them. *target*s must
have entries in the source state. They cannot be externals.

## Examples

```sh
chezmoi forget ~/.bashrc
```

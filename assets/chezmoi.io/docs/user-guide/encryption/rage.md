# rage

chezmoi supports encrypting files with [rage](https://str4d.xyz/rage).

To use rage, set `age.command` to `rage` in your configuration file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    command = "rage"
```

Then, configure chezmoi as you would for [age](age.md).

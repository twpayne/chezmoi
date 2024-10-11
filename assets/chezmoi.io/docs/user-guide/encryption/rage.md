# rage

chezmoi supports encrypting files with [rage](https://str4d.xyz/rage).

To use rage, set `age.command` to `rage` in your configuration file, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
encryption = "age"
[age]
    command = "rage"
```

!!! note

    Make sure `encryption` is added to the top level section at the beginning of
    the config, before any other sections.

Then, configure chezmoi as you would for [age](age.md).

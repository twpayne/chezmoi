# Warnings

By default, chezmoi will warn you when it encounters potential problems. Some of
these warnings can be suppressed by setting values in configuration file.

| Variable                       | Type | Default | Description                                    |
| ------------------------------ | ---- | ------- | ---------------------------------------------- |
| `configFileTemplateHasChanged` | bool | `true`  | Warn when the config file template has changed |

!!! example

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [warnings]
        configFileTemplateHasChanged = false
    ```

# `.chezmoi.<format>.tmpl`

If a file called `.chezmoi.<format>.tmpl` exists then `chezmoi init` will use
it to create an initial config file. `<format>` must be one of the the
supported config file formats, e.g. `json`, `toml`, or `yaml`.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoi.yaml.tmpl"
    {{ $email := promptString "email" -}}

    data:
        email: {{ $email | quote }}
    ```

# `.chezmoi.$FORMAT.tmpl`

If a file called `.chezmoi.$FORMAT.tmpl` exists then `chezmoi init` will use it
to create an initial config file. `$FORMAT` must be one of the supported
config file formats, e.g. `json`, `toml`, or `yaml`.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoi.yaml.tmpl"
    {{ $email := promptString "email" -}}

    data:
        email: {{ $email | quote }}
    ```

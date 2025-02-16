# `.chezmoi.$FORMAT.tmpl`

If a file called `.chezmoi.$FORMAT.tmpl` exists then `chezmoi init` will use it
to create an initial config file. `$FORMAT` must be one of the supported config
file formats. Templates defined in `.chezmoitemplates` are not available because
the template is executed before the source state is read.

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoi.yaml.tmpl"
    {{ $email := promptString "email" -}}

    data:
        email: {{ $email | quote }}
    ```

--8<-- "config-format.md"

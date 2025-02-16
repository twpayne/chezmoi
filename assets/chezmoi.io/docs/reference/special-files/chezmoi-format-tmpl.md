# `.chezmoi.$FORMAT.tmpl`

If a file called `.chezmoi.$FORMAT.tmpl` exists in the root of the source state
then [`chezmoi init`][init] will use it to create or update the chezmoi config
file. `$FORMAT` must be one of the supported config file formats.

This template differs from source state templates because this template is
executed prior to the reading of the source state.

| Feature                                            | Available? |
| -------------------------------------------------- | ---------- |
| data in the [config file][config]                  | ✅         |
| data in [`.chezmoidata.$FORMAT`][data-files] files | 🚫         |
| data in [`.chezmoidata/`][data-dirs] directories   | 🚫         |
| templates in [`.chezmoitemplates`][templates]      | 🚫         |
| [template functions][functions]                    | ✅         |
| [init functions][init-functions]                   | ✅         |

!!! example

    ``` title="~/.local/share/chezmoi/.chezmoi.yaml.tmpl"
    {{ $email := promptStringOnce . "email" "What is your email address" -}}

    data:
        email: {{ $email | quote }}
    ```

--8<-- "config-format.md"

!!! info

    This file will also be used to update the config file when a command
    supports the `--init` flag, such as `chezmoi update --init`.

[config]: /reference/configuration-file/index.md
[data-dirs]: /reference/special-directories/chezmoidata.md
[data-files]: /reference/special-files/chezmoidata-format.md
[functions]: /reference/templates/functions/index.md
[init-functions]: /reference/templates/init-functions/index.md
[init]: /reference/commands/init.md
[templates]: /reference/special-directories/chezmoitemplates.md

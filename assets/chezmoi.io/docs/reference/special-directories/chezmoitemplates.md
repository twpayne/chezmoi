# `.chezmoitemplates/`

If a directory called `.chezmoitemplates/` exists in the root of the source
directory, then all files in this directory are available as templates with a
name equal to the relative path to the `.chezmoitemplates/` directory.

The [`template` action][action] or [`includeTemplate` function][function] can be
used to include these templates in another template. The context value (`.`)
must be set explicitly if needed, otherwise the template will be executed with
`nil` context data.

!!! example

    Given:

    ``` title="~/.local/share/chezmoi/.chezmoitemplates/foo"
    {{ if true }}bar{{ end }}
    ```

    ``` title="~/.local/share/chezmoi/dot_file.tmpl"
    {{ template "foo" . }}
    ```

    The target state of `.file` will be `bar`.

[action]: https://pkg.go.dev/text/template#hdr-Actions
[function]: /reference/templates/functions/includeTemplate.md

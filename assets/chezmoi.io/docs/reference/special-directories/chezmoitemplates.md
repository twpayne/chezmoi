# `.chezmoitemplates`

If a directory called `.chezmoitemplates` exists, then all files in this
directory are available as templates with a name equal to the relative path
to the `.chezmoitemplates` directory.

The [`template` action](https://pkg.go.dev/text/template#hdr-Actions) can be
used to include these templates in another template. The value of `.` must be
set explicitly if needed, otherwise the template will be executed with `nil`
data.

!!! example

    Given:

    ``` title="~/.local/share/chezmoi/.chezmoitemplates/foo"
    {{ if true }}bar{{ end }}
    ```

    ``` title="~/.local/share/chezmoi/dot_file.tmpl"
    {{ template "foo" . }}
    ```

    The target state of `.file` will be `bar`.

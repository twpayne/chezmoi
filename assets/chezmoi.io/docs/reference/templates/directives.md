# Directives

File-specific template options can be set using template directives in the
template of the form:

    chezmoi:template:$KEY=$VALUE

which sets the template option `$KEY` to `$VALUE`. `$VALUE` must be quoted if it
contains spaces or double quotes. Multiple key/value pairs may be specified on a
single line.

Lines containing template directives are removed to avoid parse errors from any
delimiters. If multiple directives are present in a file, later directives
override earlier ones.

## Delimiters

By default, chezmoi uses the standard `text/template` delimiters `{{` and `}}`.
If a template contains the string:

    chezmoi:template:left-delimiter=$LEFT right-delimiter=$RIGHT

Then the delimiters `$LEFT` and `$RIGHT` are used instead. Either or both of
`left-delimiter=$LEFT` and `right-delimiter=$RIGHT` may be omitted. If either
`$LEFT` or `$RIGHT` is empty then the default delimiter (`{{` and `}}`
respectively) is set instead.

The delimiters are specific to the file in which they appear and are not
inherited by templates called from the file.

!!! example

    ```sh
    #!/bin/sh
    # chezmoi:template:left-delimiter="# [[" right-delimiter=]]

    # [[ "true" ]]
    ```

## Format indent

By default, chezmoi's `toJson`, `toToml`, and `toYaml` template functions use
the default indent of two spaces. The indent can be overidden with:

    chezmoi:template:format-indent=$STRING

to set the indent to be the literal `$STRING`, or

    chezmoi:template:format-indent-width=$WIDTH

to set the indent to be `$WIDTH` spaces.

!!! example

    ```
    {{/* chezmoi:template:format-indent="\t" */}}
    {{ dict "key" "value" | toJson }}
    ```

!!! example

    ```
    {{/* chezmoi:template:format-indent-width=4 */}}
    {{ dict "key" "value" | toYaml }}
    ```

## Line endings

Many of the template functions available in chezmoi primarily use UNIX-style
line endings (`lf`/`\n`), which may result in unexpected output when running
`chezmoi diff` on a `modify_` template. These line endings can be overridden
with a template directive:

    chezmoi:template:line-endings=$VALUE

`$VALUE` can be an arbitrary string or one of:

| Value    | Effect                                                               |
| -------- | -------------------------------------------------------------------- |
| `crlf`   | Use Windows line endings (`\r\n`)                                    |
| `lf`     | Use UNIX-style line endings (`\n`)                                   |
| `native` | Use platform-native line endings (`crlf` on Windows, `lf` elsewhere) |

## Missing keys

By default, chezmoi will return an error if a template indexes a map with a key
that is not present in the map. This behavior can be changed globally with the
`template.options` configuration variable or with a template directive:

    chezmoi:template:missing-key=$VALUE

`$VALUE` can be one of:

| Value     | Effect                                                                                        |
| --------- | --------------------------------------------------------------------------------------------- |
| `error`   | Return an error on any missing key (default)                                                  |
| `invalid` | Ignore missing keys. If printed, the result of the index operation is the string `<no value>` |
| `zero`    | Ignore missing keys. If printed, the result of the index operation is the zero value          |

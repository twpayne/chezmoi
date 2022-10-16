# Delimiters

By default, chezmoi uses the standard `text/template` delimiters `{{` and `}}`.
If a template contains the string

    chezmoi:template:left-delimiter=$LEFT right-delimiter=$RIGHT

Then the delimiters `$LEFT` and `$RIGHT` are used instead. Either or both of
`left-delimiter=$LEFT` and `right-delimiter=$RIGHT` may be omitted. `$LEFT` and
`$RIGHT` must be quoted if they contain spaces. If either `$LEFT` or `$RIGHT`
is empty then the default delimiter (`{{` and `}}` respectively) is set
instead.

chezmoi will remove the line containing the `chezmoi:template:` directive to
avoid parse errors from the delimiters. Only the first encountered directive is
considered.

The delimiters are specific to the file in which they appear and are not
inherited by templates called from the file.

!!! example

    ```sh
    #!/bin/sh
    # chezmoi:template:left-delimiter="# [[" right-delimiter=]]

    # [[ "true" ]]
    ```

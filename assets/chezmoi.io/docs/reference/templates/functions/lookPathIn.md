# `lookPathIn` *file* *paths*

`lookPathIn` searches for an executable named *file* in the directories provided by
the `paths` parameter using the standard OS way of separating the PATH environment
variable. The result may be an absolute path or a path relative to the current directory.
If *file* is not found, `lookPathIn` returns an empty string.

If the OS is Windows `lookPathIn` will either: if there is an extension, check to see if
the extension is specified in the `PathExt` environment variable. If there isn't an
extension it will try each of the extensions specified in the `PathExt` environment
variable in the order provided until it finds one. In either case if it doesn't `lookPathIn`
moves onto the next path provided in the `paths` parameter.

`lookPathIn` is provided as an alternative to `lookPath` so that you interrogate the
paths as you would have them.

Each successful lookup is cached based on the full path, and evaluated in the correct
order each time to reduce `File Stat` operations.

!!! example

    ```
        {{- $paths := list }}
        {{- $homeDir := .chezmoi.homeDir }}
        {{- range $_, $relPath := list "bin" "go/bin" ".cargo/bin" ".local/bin" }}
        {{    $path := joinPath $homeDir $relPath }}
        {{-   if stat $path }}
        {{-     $paths = mustAppend $paths $path }}
        {{-   end }}
        {{- end }}
        {{- if $paths }}
        export PATH={{ toStrings $paths | join ":" }}:$PATH
        {{- end }}

        {{ if lookPath "less" $paths }}
        echo "Good news we have found 'less' on system at '{{ lookPath "less" $paths }}'!"
        export DIFFTOOL=less
        {{ end }}
    ```

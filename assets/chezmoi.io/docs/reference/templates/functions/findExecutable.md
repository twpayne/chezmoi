# `findExecutable` *file* *...paths*

`findExecutable` searches for an executable named *file* in the directories provided by
the varargs `paths` parameter. In the case of Windows it will return the correct extension
if none was provided based on `PathExt`.

The input to `findExecutable` is flattened recursively. It can be a variable parameter array such as:
```
{{ findExecutable "less" "bin" "go/bin" ".cargo/bin" ".local/bin" }}
```
Or a slice:
```
{{ findExecutable "less" (list "bin" "go/bin" ".cargo/bin" ".local/bin") }}
```

`findExecutable` is provided as an alternative to `lookPath` so that you interrogate the
paths as you would have them after deployment of the RC script.

Each successful lookup is cached based on the full path, and evaluated in the correct
order each time to reduce `File` `Stat` operations.

!!! example

    ```
        {{ if findExecutable "less" "bin" "go/bin" ".cargo/bin" ".local/bin" }}
        echo "Good news we have found 'less' on system at '{{ findExecutable "less" "bin" "go/bin" ".cargo/bin" ".local/bin" }}'!"
        export DIFFTOOL=less
        {{ end }}
    ```

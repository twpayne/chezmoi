# `replaceAllRegex` *expr* *repl* *text*

`replaceAllRegex` returns *text* with all substrings matching the regular
expression *expr* replaced with *repl*. It is an alternative to [sprig's
`regexpReplaceAll` function][sprig] with a different argument order that
supports pipelining.

!!! example

    ```
    {{ "foo subject string" | replaceAllRegex "foo" "bar" }}
    ```

[sprig]: http://masterminds.github.io/sprig/strings.html

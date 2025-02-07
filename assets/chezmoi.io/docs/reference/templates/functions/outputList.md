# `outputList` *name* [*argList*]

`outputList` returns the output of executing the command *name* with the
*argList*. If executing the command returns an error then template execution
exits with an error. The execution occurs every time that the template is
executed. It is the user's responsibility to ensure that executing the command
is both idempotent and fast.

This differs from [`output`][output] in that it allows for the *argsList* to be
created programmatically.

!!! example

    ```
    {{- $args := (list "config" "current-context")  }}
    current-context: {{ outputList "kubectl" $args | trim }}
    ```

[output]: /reference/templates/functions/output.md

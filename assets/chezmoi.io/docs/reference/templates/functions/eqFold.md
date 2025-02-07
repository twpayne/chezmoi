# `eqFold` *string1* *string2* [*extraString*...]

`eqFold` returns the boolean truth of comparing *string1* with *string2* and
any number of *extraString*s under Unicode case-folding.

!!! example

    ```
    {{ $commandOutput := output "path/to/output-FOO.sh" }}
    {{ if eqFold "foo" $commandOutput }}
    # $commandOutput is "foo"/"Foo"/"FOO"...
    {{ else if eqFold "bar" $commandOutput }}
    # $commandOutput is "bar"/"Bar"/"BAR"...
    {{ end }}
    ```

+++ 2.25.0

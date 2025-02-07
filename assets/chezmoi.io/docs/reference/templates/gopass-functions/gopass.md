# `gopass` *gopass-name*

`gopass` returns passwords stored in [gopass][gopass] using the gopass CLI
(`gopass`). *gopass-name* is passed to `gopass show --password $GOPASS_NAME` and
the first line of the output of `gopass` is returned with the trailing newline
stripped. The output from `gopass` is cached so calling `gopass` multiple times
with the same *gopass-name* will only invoke `gopass` once.

!!! example

    ```
    {{ gopass "$PASS_NAME" }}
    ```

[gopass]: https://www.gopass.pw/

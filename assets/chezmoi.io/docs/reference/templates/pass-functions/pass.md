# `pass` *pass-name*

`pass` returns passwords stored in [pass][pass] using the pass CLI (`pass`).
*pass-name* is passed to `pass show $PASS_NAME` and the first line of the output
of `pass` is returned with the trailing newline stripped. The output from `pass`
is cached so calling `pass` multiple times with the same *pass-name* will only
invoke `pass` once.

!!! example

    ```
    {{ pass "$PASS_NAME" }}
    ```

[pass]: https://www.passwordstore.org/

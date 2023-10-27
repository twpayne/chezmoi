# `passage` *pass-name*

`passage` returns passwords stored in
[passage](https://github.com/FiloSottile/passage) using the passage CLI
(`passage`). *pass-name* is passed to `passage show $PASS_NAME` and the first
line of the output of `passage` is returned with the trailing newline
stripped. The output from `passage` is cached so calling `passage` multiple
times with the same *pass-name* will only invoke `passage` once.

!!! example

    ```
    {{ passage "$PASS_NAME" }}
    ```


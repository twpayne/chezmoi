# `pass` *pass-name*

`pass` returns passwords stored in [pass](https://www.passwordstore.org/) using
the pass CLI (`pass`). *pass-name* is passed to `pass show <pass-name>` and the
first line of the output of `pass` is returned with the trailing newline
stripped. The output from `pass` is cached so calling `pass` multiple times
with the same *pass-name* will only invoke `pass` once.

!!! example

    ```
    {{ pass "<pass-name>" }}
    ```

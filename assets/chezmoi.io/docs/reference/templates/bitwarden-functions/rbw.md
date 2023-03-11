# `rbw` [*arg*...]

`rbw` returns structured data retrieved from [Bitwarden](https://bitwarden.com)
using [`rbw`](https://github.com/doy/rbw). *arg*s are passed to `rbw get --raw`
and the output is parsed as JSON.

The output from `rbw get --raw` is cached so calling `rbw` multiple times with
the same arguments will only invoke `rbw` once.

!!! example

    ```
    username = {{ (rbw "test-entry").data.username }}
    password = {{ (rbw "test-entry").data.password }}
    ```

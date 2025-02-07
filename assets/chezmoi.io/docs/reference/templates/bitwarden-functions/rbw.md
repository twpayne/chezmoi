# `rbw` *name* [*arg*...]

`rbw` returns structured data retrieved from [Bitwarden][bitwarden] using
[`rbw`][rbw]. *name* is passed to `rbw get --raw`, along with any extra *arg*s,
and the output is parsed as JSON.

The output from `rbw get --raw` is cached so calling `rbw` multiple times with
the same arguments will only invoke `rbw` once.

!!! example

    ```
    username = {{ (rbw "test-entry").data.username }}
    password = {{ (rbw "test-entry" "--folder" "my-folder").data.password }}
    ```

[bitwarden]: https://bitwarden.com
[rbw]: https://github.com/doy/rbw

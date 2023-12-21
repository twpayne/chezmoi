# `rbwFields` *name* [*arg*...]

`rbw` returns structured data retrieved from [Bitwarden](https://bitwarden.com)
using [`rbw`](https://github.com/doy/rbw). *name* is passed to `rbw get --raw`,
along with any extra *arg*s, the output is parsed as JSON, and the elements
of `fields` are returned as a dict indexed by each field's `name`.

The output from `rbw get --raw` is cached so calling `rbwFields` multiple times with
the same arguments will only invoke `rbwFields` once.

!!! example

    ```
    {{ (rbwFields "item").name.value }}
    {{ (rbwFields "item" "--folder" "my-folder").name.value }}
    ```

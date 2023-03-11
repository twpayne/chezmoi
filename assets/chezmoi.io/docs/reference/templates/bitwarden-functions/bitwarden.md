# `bitwarden` [*arg*...]

`bitwarden` returns structured data retrieved from
[Bitwarden](https://bitwarden.com) using the [Bitwarden
CLI](https://bitwarden.com/help/cli) (`bw`). *arg*s are passed to `bw get`
unchanged and the output from `bw get` is parsed as JSON.

The output from `bw get` is cached so calling `bitwarden` multiple times with
the same arguments will only invoke `bw` once.

!!! example

    ```
    username = {{ (bitwarden "item" "$ITEMID").login.username }}
    password = {{ (bitwarden "item" "$ITEMID").login.password }}
    ```

# `keepassxcAttachment` *entry* *name*

`keepassxcAttachment` returns the attachment with *name* of *entry* using
`keepassxc-cli`.

!!! info

    `keepassxcAttachment` requires `keepassxc-cli` version 2.7.0 or later.

!!! example

    ```
    {{- keepassxcAttachment "SSH Config" "config" -}}
    ```

+++ 2.15.0

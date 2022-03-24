# `keepassxcAttachment` *entry* *name*

`keepassxcAttachment` returns the attachment with *name* of *entry* using
`keepassxc-cli`. It behaves identically to the `keepassxc` function in terms of
configuration, password prompting, password storage, and result caching.

!!! example

    ```
    {{- keepassxcAttachment "SSH Config" "config" -}}
    ```

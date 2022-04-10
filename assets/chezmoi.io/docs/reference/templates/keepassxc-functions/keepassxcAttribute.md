# `keepassxcAttribute` *entry* *attribute*

`keepassxcAttribute` returns the attribute *attribute* of *entry* using
`keepassxc-cli`, with any leading or trailing whitespace removed.

!!! example

    ```
    {{ keepassxcAttribute "SSH Key" "private-key" }}
    ```

# `keepassxcAttribute` *entry* *attribute*

`keepassxcAttribute` returns the attribute *attribute* of *entry* using
`keepassxc-cli`, with any leading or trailing whitespace removed. It behaves
identically to the `keepassxc` function in terms of configuration, password
prompting, password storage, and result caching.

!!! example

    ```
    {{ keepassxcAttribute "SSH Key" "private-key" }}
    ```

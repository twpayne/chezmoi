# `protonPassJSON` *uri*

`protonPassJSON` returns the structured data associated with *uri* from [Proton
Pass][protonpass] using the [Proton Pass CLI][protonpass-cli].

!!! example

    ```
    {{ (protonPassJSON "pass://$SHARE_ID/$ITEM_ID").item.content.content.key.password }}
    ```

[protonpass]: https://proton.me/pass
[protonpass-cli]: https://protonpass.github.io/pass-cli/

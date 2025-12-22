# `protonPass` *uri*

`protonPass` returns the item associated with *uri* from [Proton
Pass][protonpass] using the [Proton Pass CLI][protonpass-cli].

!!! example

    ```
    {{ protonPass "pass://$SHARE_ID/$ITEM_ID/$FIELD" }}
    ```

[protonpass]: https://proton.me/pass
[protonpass-cli]: https://protonpass.github.io/pass-cli/

# `protonPassAttachment` *share-id* *item-id* *attachment-id*

`protonPassAttachment` returns the content of the given attachment from [Proton
Pass][protonpass] using the [Proton Pass CLI][protonpass-cli]. The output of
`pass-cli` is cached so calling `protonPassAttachment` multiple times with the
same *share-id*, *item-id*, and *attachment-id* will only invoke `pass-cli`
once.

!!! example

    ```
    {{ protonPassAttachment "$SHARE_ID" "$ITEM_ID" "$ATTACHMENT_ID" }}
    ```

[protonpass]: https://proton.me/pass
[protonpass-cli]: https://protonpass.github.io/pass-cli/

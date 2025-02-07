# `bitwardenAttachment` *filename* *itemid*

`bitwardenAttachment` returns a document from [Bitwarden][bitwarden] using the
[Bitwarden CLI][cli] (`bw`). *filename* and *itemid* are passed to `bw get
attachment $FILENAME --itemid $ITEMID` and the output is returned.

The output from `bw` is cached so calling `bitwardenAttachment` multiple times
with the same *filename* and *itemid* will only invoke `bw` once.

!!! example

    ```
    {{- bitwardenAttachment "$FILENAME" "$ITEMID" -}}
    ```

[bitwarden]: https://bitwarden.com/
[cli]: https://bitwarden.com/help/article/cli/

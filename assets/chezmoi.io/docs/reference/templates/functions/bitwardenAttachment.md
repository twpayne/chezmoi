# `bitwardenAttachment` *filename* *itemid*

`bitwardenAttachment` returns a document from
[Bitwarden](https://bitwarden.com/) using the [Bitwarden
CLI](https://bitwarden.com/help/article/cli/) (`bw`). *filename* and *itemid*
is passed to `bw get attachment $FILENAME --itemid $ITEMID` and the output from
`bw` is returned. The output from `bw` is cached so calling
`bitwardenAttachment` multiple times with the same *filename* and *itemid* will
only invoke `bw` once.

!!! example

    ```
    {{- bitwardenAttachment "$FILENAME" "$ITEMID" -}}
    ```

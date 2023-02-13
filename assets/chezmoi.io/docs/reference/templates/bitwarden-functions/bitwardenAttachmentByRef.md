# `bitwardenAttachmentByRef` *filename* *args*

`bitwardenAttachmentByRef` returns a document from
[Bitwarden](https://bitwarden.com/) using the [Bitwarden
CLI](https://bitwarden.com/help/article/cli/) (`bw`).  This method requires two
calls to `bw` to complete.  *args* are passed to `bw get` in order to retrieve
the item's *itemid*.  Then, *filename* and *itemid* are passed to
`bw get attachment $FILENAME --itemid $ITEMID` and the output from
`bw` is returned. The output from `bw` is cached so calling
`bitwardenAttachmentByRef` multiple times with the same *filename* and *itemid* will
only invoke `bw` once.

!!! example

    ```
    {{- bitwardenAttachmentByRef "$FILENAME" "$ARGS" -}}
    ```

!!! example

    ```
    {{- bitwardenAttachmentByRef "id_rsa" "item" "example.com" -}}
    ```

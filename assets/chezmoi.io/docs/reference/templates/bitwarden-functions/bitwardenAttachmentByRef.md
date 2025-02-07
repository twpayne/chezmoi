# `bitwardenAttachmentByRef` *filename* *args*

`bitwardenAttachmentByRef` returns a document from [Bitwarden][bitwarden] using
the [Bitwarden CLI][cli] (`bw`). This method requires two calls to `bw` to
complete:

1. First, *args* are passed to `bw get` in order to retrieve the item's
   *itemid*.
2. Then, *filename* and *itemid* are passed to `bw get attachment $FILENAME
   --itemid $ITEMID` and the output from `bw` is returned.

The output from `bw` is cached so calling `bitwardenAttachmentByRef` multiple
times with the same *filename* and *itemid* will only invoke `bw` once.

!!! example

    ```
    {{- bitwardenAttachmentByRef "$FILENAME" "$ARGS" -}}
    ```

!!! example

    ```
    {{- bitwardenAttachmentByRef "id_rsa" "item" "example.com" -}}
    ```

[bitwarden]: https://bitwarden.com/
[cli]: https://bitwarden.com/help/article/cli/

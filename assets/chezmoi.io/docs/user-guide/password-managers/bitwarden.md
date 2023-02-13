# Bitwarden

chezmoi includes support for [Bitwarden](https://bitwarden.com/) using the
[Bitwarden CLI](https://bitwarden.com/help/cli) to expose data as a template
function.

Log in to Bitwarden using:

```console
$ bw login $BITWARDEN_EMAIL
```

Unlock your Bitwarden vault:

```console
$ bw unlock
```

Set the `BW_SESSION` environment variable, as instructed.

!!! tip "Bitwarden Session One-liner"

    Set `BW_SESSION` automatically with:

    ```console
    $ export BW_SESSION=$(bw {login,unlock} --raw)
    ```

The structured data from `bw get` is available as the `bitwarden` template
function in your config files, for example:

```
username = {{ (bitwarden "item" "example.com").login.username }}
password = {{ (bitwarden "item" "example.com").login.password }}
```

Custom fields can be accessed with the `bitwardenFields` template function. For
example, if you have a custom field named `token` you can retrieve its value
with:

```
{{ (bitwardenFields "item" "example.com").token.value }}
```

Attachments can be accessed with the `bitwardenAttachment` and
`bitwardenAttachmentByRef` template function. For example, if you have an
attachment named `id_rsa`, you can retrieve its value with:

```
{{ bitwardenAttachment "id_rsa" "bf22e4b4-ae4a-4d1c-8c98-ac620004b628" }}
```
or
```
{{ bitwardenAttachmentByRef "id_rsa" "item" "example.com" }}
```

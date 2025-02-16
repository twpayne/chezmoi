# Bitwarden

chezmoi includes support for [Bitwarden][bitwarden] using the [Bitwarden
CLI][cli] (`bw`), [Bitwarden Secrets CLI][secrets] (`bws`), and [`rbw`][rbw]
commands to expose data as a template function.

## Bitwarden CLI

Log in to Bitwarden using a normal method

```sh
bw login $BITWARDEN_EMAIL # or
bw login --apikey # or
bw login --sso
```

If required, unlock your Bitwarden vault (API key and SSO logins always require
an explicit unlock step):

```sh
bw unlock
```

Set the `BW_SESSION` environment variable, as instructed.

!!! tip "Bitwarden Session One-liner"

    The `BW_SESSION` value can be set directly. The exact combination differs
    based on whether you are currently logged into Bitwarden and how you log
    into Bitwarden.

    ```sh
    export BW_SESSION=$(bw unlock --raw)                    # You are already logged in with any method
    export BW_SESSION=$(bw login $BITWARDEN_EMAIL --raw)    # You are not logged in and log in with an email
    export BW_SESSION=$(bw login --sso && bw unlock --raw)  # You are not logged in and login with SSO or API key
    ```

The structured data from `bw get` is available as the `bitwarden` template
function in your config files, for example:

```text
username = {{ (bitwarden "item" "example.com").login.username }}
password = {{ (bitwarden "item" "example.com").login.password }}
```

Custom fields can be accessed with the `bitwardenFields` template function. For
example, if you have a custom field named `token` you can retrieve its value
with:

```text
{{ (bitwardenFields "item" "example.com").token.value }}
```

Attachments can be accessed with the `bitwardenAttachment` and
`bitwardenAttachmentByRef` template function. For example, if you have an
attachment named `id_rsa`, you can retrieve its value with:

```text
{{ bitwardenAttachment "id_rsa" "bf22e4b4-ae4a-4d1c-8c98-ac620004b628" }}
```

or

```text
{{ bitwardenAttachmentByRef "id_rsa" "item" "example.com" }}
```

## Bitwarden Secrets CLI

Generate an [access token][token] for a specific [service account][account].

Either set the `BWS_ACCESS_TOKEN` environment variable or store the access token
in a template variable, e.g.

```toml title="~/.config/chezmoi/chezmoi.toml"
[data]
    accessToken = "0.48c78342-1635-48a6-accd-afbe01336365.C0tMmQqHnAp1h0gL8bngprlPOYutt0:B3h5D+YgLvFiQhWkIq6Bow=="
```

You can then retrieve secrets using the `bitwardenSecrets` template function, for
example:

```text
{{ (bitwardenSecrets "be8e0ad8-d545-4017-a55a-b02f014d4158" .accessToken).value }}
```

[bitwarden]: https://bitwarden.com/
[cli]: https://bitwarden.com/help/cli
[secrets]: https://bitwarden.com/help/secrets-manager-cli/
[rbw]: https://github.com/doy/rbw
[token]: https://bitwarden.com/help/access-tokens/
[account]: https://bitwarden.com/help/service-accounts/

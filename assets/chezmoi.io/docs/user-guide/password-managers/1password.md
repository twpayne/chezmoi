# 1Password

chezmoi includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

!!! note

    [1Password CLI 2.0](https://developer.1password.com/) has been released.
    Examples will be shown using the changed details for this version and
    examples for 1Password CLI 1.x will follow.

Log in and get a session using:

```console
# For 1Password 2.x. Neither step is necessary with biometric authentication.
$ op account add --address $SUBDOMAIN.1password.com --email $EMAIL
$ eval $(op signin --account $SUBDOMAIN)
```

??? info

    ```console
    # For 1Password 1.x
    $ eval $(op signin $SUBDOMAIN.1password.com $EMAIL)
    ```

The output of `op item get $UUID--format json` (`op get item $UUID`) is
available as the `onepassword` template function. chezmoi parses the JSON output
and returns it as structured data. For example, if the output is:

```json
{
  "id": "$UUID",
  "title": "$TITLE",
  "version": 2,
  "vault": {
    "id": "$vaultUUID"
  },
  "category": "LOGIN",
  "last_edited_by": "$userUUID",
  "created_at": "2010-08-23T13:18:43Z",
  "updated_at": "2014-07-20T04:40:11Z",
  "fields": [
    {
      "id": "username",
      "type": "STRING",
      "purpose": "USERNAME",
      "label": "username",
      "value": "$USERNAME"
    },
    {
      "id": "password",
      "type": "CONCEALED",
      "purpose": "PASSWORD",
      "label": "password",
      "value": "$PASSWORD",
      "password_details": {
        "strength": "FANTASTIC",
        "history": []
      }
    }
  ],
  "urls": [
    {
      "primary": true,
      "href": "$URL"
    }
  ]
}
```

Then you can access the password field with the syntax

```
{{ (onepassword "$UUID").fields[1].value }}
```

or:

```
{{ range (onepassword "$UUID").fields -}}
{{- if and (eq .label "password") (eq .purpose "PASSWORD") }}{{ .value }}{{ end -}}
{{- end }}
```

??? info

    1Password CLI 1.x returns a simpler structure:

    ```json
    {
      "uuid": "$UUID",
      "details": {
        "password": "$PASSWORD"
      }
    }
    ```

    This allows for the syntax:

    ```
    {{ (onepassword "$UUID").details.password }}
    ```

`onepasswordDetailsFields` returns a reworked version of the structure that
allows the fields to be queried by key:

```json
{
  "password": {
    "id": "password",
    "label": "password",
    "password_details": {
      "history": [],
      "strength": "FANTASTIC"
    },
    "purpose": "PASSWORD",
    "type": "CONCEALED",
    "value": "$PASSWORD"
  },
  "username": {
    "id": "username",
    "label": "username",
    "purpose": "USERNAME",
    "type": "STRING",
    "value": "$USERNAME"
  }
}
```

```
{{- (onepasswordDetailsFields "$UUID").password.value }}
```

Additional fields may be obtained with `onePasswordItemFields`; not all objects
in 1Password have item fields, so it is worth testing before using:

```console
chezmoi execute-template "{{- onepasswordItemFields \"$UUID\" | toJson -}}" | jq .
```

Documents can be retrieved with:

```
{{- onepasswordDocument "$UUID" -}}
```

!!! note

    The extra `-` after the opening `{{` and before the closing `}}` instructs
    the template language to remove any whitespace before and after the
    substitution. This removes any trailing newline added by your editor when
    saving the template.

## 1Password sign-in prompt

chezmoi will verify the availability and validity of a session token in the
current environment. If it is missing or expired, you will be interactively
prompted to sign-in again.

In the past chezmoi used to simply exit with an error when no valid session was
available. If you'd like to restore that behavior, set the following option in
your configuration file:

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    prompt = false
```

!!! danger

    Do not use the prompt on shared machines. A session token verified or
    acquired interactively will be passed to the 1Password CLI through a command
    line parameter, which is visible to other users of the same system.

!!! info

    If you're using [1Password CLI
    2.0](https://developer.1password.com/docs/cli/), then the structure of the
    data returned by the `onepassword`, `onepasswordDetailsFields`, and
    `onePasswordItemFiles` template functions is different and templates will
    need to be updated.

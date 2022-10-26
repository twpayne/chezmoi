# 1Password

chezmoi includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

!!! note

    The [1Password CLI 2.0](https://developer.1password.com/) has been released.
    Examples will be shown using the changed details for this version and
    examples for 1Password CLI 1.x will follow.

Log in and get a session using:

```console
$ op account add --address $SUBDOMAIN.1password.com --email $EMAIL
$ eval $(op signin --account $SUBDOMAIN)
```

This is not necessary if you are using biometric authentication.

!!! info

    For 1Password CLI 1.x, use:

    ```console
    $ eval $(op signin $SUBDOMAIN.1password.com $EMAIL)
    ```

The output of `op read $URL` is available as the `onepasswordRead` template
function, for example:

```
{{ onepasswordRead "op://app-prod/db/password" }}
```

returns the output of

```console
$ op read op://app-prod/db/password
```

Documents can be retrieved with:

```
{{- onepasswordDocument "$UUID" -}}
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
{{ (index (onepassword "$UUID").fields 1).value }}
```

or:

```
{{ range (onepassword "$UUID").fields -}}
{{   if and (eq .label "password") (eq .purpose "PASSWORD") -}}
{{     .value -}}
{{   end -}}
{{ end }}
```

!!! info

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
in 1Password have item fields. This can be tested with:

```console
$ chezmoi execute-template "{{ onepasswordItemFields \"$UUID\" | toJson }}" | jq .
```

!!! note

    The extra `-` after the opening `{{` and before the closing `}}` instructs
    the template language to remove any whitespace before and after the
    substitution. This removes any trailing newline added by your editor when
    saving the template.

## Sign-in prompt

chezmoi will verify the availability and validity of a session token in the
current environment. If it is missing or expired, you will be interactively
prompted to sign-in again.

In the past chezmoi used to simply exit with an error when no valid session was
available. If you'd like to restore this behavior, set the `onepassword.prompt`
configuration variable to `false`, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    prompt = false
```

!!! danger

    Do not use prompt on shared machines. A session token verified or acquired
    interactively will be passed to the 1Password CLI through a command line
    parameter, which is visible to other users of the same system.

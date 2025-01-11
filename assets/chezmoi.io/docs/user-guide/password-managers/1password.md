# 1Password

chezmoi includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

Log in and get a session using:

```sh
op account add --address $SUBDOMAIN.1password.com --email $EMAIL
eval $(op signin --account $SUBDOMAIN)
```

This is not necessary if you are using biometric authentication.

The output of `op read $URL` is available as the `onepasswordRead` template
function, for example:

```text
{{ onepasswordRead "op://app-prod/db/password" }}
```

returns the output of

```sh
op read op://app-prod/db/password
```

Documents can be retrieved with:

```text
{{- onepasswordDocument "$UUID" -}}
```

The output of `op item get $UUID --format json` is available as the
`onepassword` template function. chezmoi parses the JSON output and returns it
as structured data. For example, if the output is:

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

```text
{{ (index (onepassword "$UUID").fields 1).value }}
```

or:

```text
{{ range (onepassword "$UUID").fields -}}
{{   if and (eq .label "password") (eq .purpose "PASSWORD") -}}
{{     .value -}}
{{   end -}}
{{ end }}
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

```text
{{- (onepasswordDetailsFields "$UUID").password.value }}
```

Additional fields may be obtained with `onepasswordItemFields`; not all objects
in 1Password have item fields. This can be tested with:

```sh
chezmoi execute-template "{{ onepasswordItemFields \"$UUID\" | toJson }}" | jq .
```

## Sign-in prompt

chezmoi will verify the availability and validity of a session token in the
current environment. If it is missing or expired, you will be interactively
prompted to sign-in again.

In the past chezmoi used to exit with an error when no valid session was
available. If you'd like to restore this behavior, set the `onepassword.prompt`
configuration variable to `false`, for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    prompt = false
```

!!! danger

    Do not use `prompt` on shared machines. A session token verified or acquired
    interactively will be passed to the 1Password CLI through a command line
    parameter, which is visible to other users of the same system.

## Secrets Automation

chezmoi has experimental support for secrets automation with [1Password
Connect](https://developer.1password.com/docs/connect/) and [1Password Service
Accounts](https://developer.1password.com/docs/service-accounts). These might be
used on restricted machines where you cannot or do not wish to install a full
1Password desktop application.

When these features are used, the behavior of the 1Password CLI changes, so
chezmoi requires explicit configuration for either connect or service account
modes using the `onepassword.mode` configuration option. The default, if not
specified, is `account`:

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    mode = "account"
```

In `account` mode, chezmoi will stop with an error if the environment variable
`OP_SERVICE_ACCOUNT_TOKEN` is set, or if both environment variables
`OP_CONNECT_HOST` and `OP_CONNECT_TOKEN` are set.

!!! info

    Both 1Password Connect and Service Accounts prevent the CLI from working
    with multiple accounts. If you need access to secrets from more than one
    1Password account, do not use these features with chezmoi.

### 1Password Connect

Once 1Password Connect is
[configured](https://developer.1password.com/docs/connect/connect-cli#requirements),
and `OP_CONNECT_HOST` and `OP_CONNECT_TOKEN` are properly set, set
`onepassword.mode` to `connect`.

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    mode = "connect"
```

In `connect` mode:

- the `onepasswordDocument` template function is not available,
- `account` parameters are not allowed in 1Password template functions,
- chezmoi will stop with an error if one or both of `OP_CONNECT_HOST` and
  `OP_CONNECT_TOKEN` are unset, or if `OP_SERVICE_ACCOUNT_TOKEN` is set.

### 1Password Service Accounts

Once a 1Password service account has been
[created](https://developer.1password.com/docs/service-accounts/use-with-1password-cli/#requirements)
and `OP_SERVICE_ACCOUNT_TOKEN` is properly set, set `onepassword.mode` to
`service`.

```toml title="~/.config/chezmoi/chezmoi.toml"
[onepassword]
    mode = "service"
```

In `service` mode:

- `account` parameters are not allowed in 1Password template functions,
- chezmoi will stop with an error if `OP_SERVICE_ACCOUNT_TOKEN` is unset, or if
  both of `OP_CONNECT_HOST` and `OP_CONNECT_TOKEN` are set.

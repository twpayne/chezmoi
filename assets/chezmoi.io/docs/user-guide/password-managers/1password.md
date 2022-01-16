# 1Password

chezmoi includes support for [1Password](https://1password.com/) using the
[1Password CLI](https://support.1password.com/command-line-getting-started/) to
expose data as a template function.

Log in and get a session using:

```console
$ eval $(op signin <subdomain>.1password.com <email>)
```

The output of `op get item <uuid>` is available as the `onepassword` template
function. chezmoi parses the JSON output and returns it as structured data. For
example, if the output of `op get item "<uuid>"` is:

```json
{
    "uuid": "<uuid>",
    "details": {
        "password": "xxx"
    }
}
```

Then you can access `details.password` with the syntax:

```
{{ (onepassword "<uuid>").details.password }}
```

Login details fields can be retrieved with the `onepasswordDetailsFields`
function, for example:

```
{{- (onepasswordDetailsFields "uuid").password.value }}
```

Documents can be retrieved with:

```
{{- onepasswordDocument "uuid" -}}
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

!!! warning

    Do not use the prompt on shared machines. A session token verified or
    acquired interactively will be passed to the 1Password CLI through a
    command-line parameter, which is visible to other users of the same system.

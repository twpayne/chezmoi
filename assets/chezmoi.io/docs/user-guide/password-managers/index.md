# Password Manager Integration

Using a password manager with Chezmoi enables you to maintain a public
dotfiles repository while keeping your secrets secure. Chezmoi extends its
existing [templating capabilities](../templating.md) by providing password
manager specific _template functions_ for many popular password managers.

When Chezmoi applies a template with a secret referenced from a password
manager, it will automatically fetch the secret value and insert it into the
generated destination file.

## Example: Template with Password Manager Integration

Here's a practical example of a `.zshrc.tmpl` file that retrieves an CloudFlare
API token from 1Password while maintaining other standard shell configurations:

```zsh
# set up $PATH
# …

# Cloudflare API Token retrieved from 1Password for use with flarectl
export CF_API_TOKEN='{{ onepasswordRead "op://Personal/cloudlfare-api-token/password" }}'

# set up aliases and useful functions
```

In this example, the `CF_API_TOKEN` is retrieved from a 1Password vault
named `Personal`, specifically from an item called `cloudflare-api-token` in the
`password` field.

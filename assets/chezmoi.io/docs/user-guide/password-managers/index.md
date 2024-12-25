# Password Manager Integration

Using a password manager with Chezmoi enables you to maintain a public
dotfiles repository while keeping your secrets secure. Chezmoi extends its
existing [templating capabilities](../templating.md) by providing password
manager specific _template functions_ for many popular password managers.

When Chezmoi applies a template with a secret referenced from a password
manager, it will automatically fetch the secret value and insert it into the
generated destination file.

## Example: Template with Password Manager Integration

Here's a practical example of a `.zshrc.tmpl` file that retrieves an OpenAI API
key from 1Password while maintaining other standard shell configurations:

```zsh
# set up $PATH
# …

# OpenAI API Key retrieved from 1Password
export OPENAI_API_KEY='{{ onepasswordRead "op://Personal/openai-api-key/password" }}'

# set up aliases and useful functions
```

In this example, the `OPENAI_API_KEY` is retrieved from a 1Password vault
named `Personal`, specifically from an item called `openai-api-key` in the
`password` field.

# Password Manager Integration in Chezmoi

Using a password manager with Chezmoi enables you to maintain a public dotfiles
repository while keeping your secrets secure. Chezmoi provides template functions
for many popular password managers so that your templates can render sensitive
information across multiple machines.
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
`password` field. When Chezmoi applies this template, it will automatically
fetch the current value from 1Password and insert it into the generated file.

This approach allows you to version control your dotfiles while keeping
sensitive information secure in your password manager. When you update a
secret in your password manager, the next `chezmoi apply` will automatically use
the updated value.

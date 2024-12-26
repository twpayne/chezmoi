# Password Manager Integration

This guide explains how to securely manage sensitive information in your
dotfiles using Chezmoi's password manager integration. This allows you to
maintain a public dotfiles repository while keeping your secrets protected.

## Key Requirements
* Source files must be saved as templates with the `.tmpl` extension
* A supported password manager must be installed and configured

## How It Works
  1.	Create a source `.tmpl` file containing one or more _template commands_ to fetch secrets
  2.	When you run `chezmoi apply`:
        * Chezmoi reads your templates
        * Fetches secrets from your password manager using the specified template functions
        * Renders the destination files with the secrets inserted as plain text

For more information about how to create template files or convert existing
files to templates, see the applicable [templating](/assets/chezmoi.io/docs/user-guide/templating.md) documentation.

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

In this example, the `OPENAI_API_KEY` environment variable is set by
retrieving a value from a 1Password vault named `Personal`, specifically from
an item called `openai-api-key` in the `password` field.

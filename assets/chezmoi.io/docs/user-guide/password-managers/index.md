# Password Manager Integration in Chezmoi

Using a password manager with Chezmoi enables you to maintain a public dotfiles
repository while keeping your secrets secure. Chezmoi provides template functions
for many popular password managers so that your templates can render sensitive
information across multiple machines.

## Understanding Templates

Templates are a core concept in Chezmoi that allow files to contain dynamic
content, including secrets from password managers. Instead of storing
sensitive information directly in your dotfiles, templates can reference
secrets stored safely in your password manager.

**Important:** For a dotfile to retrieve information from a password manager
during the `chezmoi apply` command, it must be configured as a template.


## Working with Template Files

There are two ways to create template files in Chezmoi:

1. Creating a new template file:

       chezmoi add --template <filename>

2. Converting an existing file to a template using the `chattr` (change
   attribute) command:

       chezmoi chattr +template <filename>

All template files must have the `.tmpl` extension for Chezmoi to process them during chezmoi apply.


## Password Manager Functions

Chezmoi supports multiple password managers through built-in functions. These
functions are wrapped in double curly braces `{{ }}` to indicate that Chezmoi
should evaluate them dynamically during chezmoi apply.

Common password manager examples:

* 1Password:

      {{ onepasswordRead "op://vault/item/field" }}

* Bitwarden:

      {{ (bitwarden "item_id").login.password }}

## Example: Template with Password Manager Integration

Here's a practical example of a `.zshrc.tmpl` file that retrieves an OpenAI API
key from 1Password while maintaining other standard shell configurations:

```
# OpenAI API Key retrieved from 1Password
export OPENAI_API_KEY='{{ onepasswordRead "op://Personal/openai-api-key/password" }}'

# Common aliases
alias ll='ls -la'
alias c='clear'
alias ..='cd ..'

# Git aliases
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'

# Path exports
export PATH=$HOME/bin:/usr/local/bin:$PATH

# Auto-completion settings
autoload -Uz compinit
compinit

# Custom functions
function mkcd() {
    mkdir -p "$1" && cd "$1"
}
```

In this example, the `OPENAI_API_KEY` is retrieved from a 1Password vault
named `Personal`, specifically from an item called `openai-api-key` in the
`password` field. When Chezmoi applies this template, it will automatically
fetch the current value from 1Password and insert it into the generated file.

This approach allows you to version control your dotfiles while keeping
sensitive information secure in your password manager. When you update a
secret in your password manager, the next `chezmoi apply` will automatically use
the updated value.

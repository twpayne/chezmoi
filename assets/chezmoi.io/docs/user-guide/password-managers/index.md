# Password Manager Integration

Using a password manager with chezmoi enables you to maintain a public dotfiles
repository while keeping your secrets secure. chezmoi extends its [templating
capabilities][templating] by providing password manager specific *template
functions* for many popular password managers.

When chezmoi applies a template with a secret referenced from a password
manager, it will automatically fetch the secret value and insert it into the
generated destination file.

!!! example

    Here's a practical example of a `.zshrc.tmpl` file that retrieves an
    CloudFlare API token from 1Password while maintaining other standard shell
    configurations:

    ```zsh
    # set up $PATH
    # …

    # Cloudflare API Token retrieved from 1Password for use with flarectl
    export CF_API_TOKEN='{{ onepasswordRead "op://Personal/cloudlfare-api-token/password" }}'

    # set up aliases and useful functions
    ```

    In this example, the `CF_API_TOKEN` is retrieved from a 1Password vault
    named `Personal`, an item called `cloudflare-api-token`, and the `password`
    field.

[templating]: /user-guide/templating.md

# `getRedirectedURL` *url*

`getRedirectedURL` returns the final URL after following any HTTP redirects
from the given *url*. If the *url* does not redirect, it returns the original
*url*.

`getRedirectedURL` is not hermetic: its return value depends on the state of
the network and the remote server at the moment the template is executed.
Exercise caution when using it in your templates.

!!! example

    ```
    {{ getRedirectedURL "https://github.com/twpayne/chezmoi/releases/latest" }}
    {{ getRedirectedURL "https://github.com/twpayne/chezmoi/raw/HEAD/README.md" }}
    {{ getRedirectedURL "https://git.io/chezmoi" }}
    ```

    This will return something like:

    ```
    https://github.com/twpayne/chezmoi/releases/tag/v2.62.7
    https://raw.githubusercontent.com/twpayne/chezmoi/aa57d1d773715e02103e87f78c58b99f9b91fc0c/README.md
    https://raw.githubusercontent.com/twpayne/chezmoi/master/assets/scripts/install.sh
    ```

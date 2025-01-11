# `joinPath` *element*...

`joinPath` joins any number of path elements into a single path, separating
them with the OS-specific path separator. Empty elements are ignored. The
result is cleaned. If the argument list is empty or all its elements are empty,
`joinPath` returns an empty string. On Windows, the result will only be a UNC
path if the first non-empty element is a UNC path.

!!! example

    ```text
    {{ joinPath .chezmoi.homeDir ".zshrc" }}
    ```

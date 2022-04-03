# `gitHubKeys` *user*

`gitHubKeys` returns *user*'s public SSH keys from GitHub using the GitHub API.
The returned value is a slice of structs with `.ID` and `.Key` fields.

!!! warning

    If you use this function to populate your `~/.ssh/authorized_keys` file
    then you potentially open SSH access to anyone who is able to modify or add
    to your GitHub public SSH keys, possibly including certain GitHub
    employees. You should not use this function on publicly-accessible machines
    and should always verify that no unwanted keys have been added, for example
    by using the `-v` / `--verbose` option when running `chezmoi apply` or
    `chezmoi update`.

!!! example

    ```
    {{ range gitHubKeys "user" }}
    {{- .Key }}
    {{ end }}
    ```

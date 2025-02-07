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

    Additionally, GitHub automatically [removes keys which haven't been used in
    the last year][timeout]. This may cause your keys to be removed from
    `~/.ssh/authorized_keys` suddenly, and without any warning or indication of
    the removal. You should provide one or more keys in plain text alongside
    this function to avoid unknowingly losing remote access to your machine.

!!! example

    ```
    {{ range gitHubKeys "user" }}
    {{- .Key }}
    {{ end }}
    ```

[timeout]: https://docs.github.com/en/authentication/troubleshooting-ssh/deleted-or-missing-ssh-keys

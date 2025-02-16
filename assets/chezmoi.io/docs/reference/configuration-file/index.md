# Configuration file

chezmoi searches for its configuration file according to the [XDG Base Directory
Specification][xdg]. The base name of the config file is `chezmoi`. If multiple
configuration file formats are present, chezmoi will report an error.

--8<-- "config-format.md"

In most installations, the config file will be read from
`$HOME/.config/chezmoi/chezmoi.$FORMAT`
(`%USERPROFILE%/.config/chezmoi/chezmoi.$FORMAT`), where `$FORMAT` is one of
`json`, `jsonc`, `toml`, or `yaml`. The config file can be set explicitly with
the `--config` command line option. By default, the format is detected based on
the extension of the config file name, but can be overridden with the
`--config-format` command line option.

## Examples

=== "JSON"

    ```json title="~/.config/chezmoi/chezmoi.json"
    {
        "sourceDir": "/home/user/.dotfiles",
        "git": {
            "autoPush": true
        }
    }
    ```

=== "JSONC"

    ```jsonc title="~/.config/chezmoi/chezmoi.jsonc"
    {
        // The chezmoi source files are stored here
        "sourceDir": "/home/user/.dotfiles",
        "git": {
            "autoPush": true
        }
    }
    ```

=== "TOML"

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    sourceDir = "/home/user/.dotfiles"
    [git]
        autoPush = true
    ```

=== "YAML"

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    sourceDir: /home/user/.dotfiles
    git:
        autoPush: true
    ```

[xdg]: https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html
[json]: https://www.json.org/json-en.html
[toml]: https://github.com/toml-lang/toml
[yaml]: https://yaml.org/

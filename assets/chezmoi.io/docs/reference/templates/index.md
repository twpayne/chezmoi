# Templates

chezmoi executes templates using [`text/template`][go-template]. The result is
treated differently depending on whether the target is a file or a symlink.

If target is a file, then:

- If the result is an empty string, then the file is removed.

- Otherwise, the target file contents are result.

If the target is a symlink, then:

- Leading and trailing whitespace are stripped from the result.

- If the result is an empty string, then the symlink is removed.

- Otherwise, the target symlink target is the result.

chezmoi executes templates using `text/template`'s `missingkey=error` option,
which means that misspelled or missing keys will raise an error. This can be
overridden by setting a list of options in the configuration file.

!!! hint

    For a full list of template options, see [`Template.Option`][option].

!!! example

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [template]
        options = ["missingkey=zero"]
    ```

[go-template]: https://pkg.go.dev/text/template
[option]: https://pkg.go.dev/text/template?tab=doc#Template.Option

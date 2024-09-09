# Diff

## Use a custom diff tool

By default, chezmoi uses a built-in diff. You can use a custom tool by setting
the `diff.command` and `diff.args` configuration variables. The elements of
`diff.args` are interpreted as templates with the variables `.Destination` and
`.Target` containing filenames of the file in the destination state and the
target state respectively. For example, to use [meld](https://meldmerge.org/),
specify:

```toml title="~/.config/chezmoi/chezmoi.toml"
[diff]
    command = "meld"
    args = ["--diff", "{{ .Destination }}", "{{ .Target }}"]
```

!!! hint

    If you generate your config file from a config file template, then you'll
    need to escape the `{{` and `}}` as `{{ "{{" }}` and `{{ "}}" }}`. That way
    your generated config file contains the `{{` and `}}` you expect.

## Use VSCode as the diff tool

To use [VSCode](https://code.visualstudio.com/) as the diff tool, add the
following to your config:

=== "TOML"

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [diff]
    command = "code"
    args = ["--wait", "--diff"]
    ```

=== "YAML"

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    diff:
      command: "code"
      args:
      - "--wait"
      - "--diff"
    ```

## Don't show scripts in the diff output

By default, `chezmoi diff` will show all changes, including the contents of
scripts that will be run. You can exclude scripts from the diff output by
setting the `diff.exclude` configuration variable in your configuration file,
for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[diff]
    exclude = ["scripts"]
```

## Don't show externals in the diff output

To exclude diffs from externals, either pass the `--exclude=externals` flag or
set `diff.exclude` to `["externals"]` in your config file.

## Customize the diff pager

You can change the diff format, and/or pipe the output into a pager of your
choice by setting `diff.pager` configuration variable. For example, to use
[`diff-so-fancy`](https://github.com/so-fancy/diff-so-fancy) specify:

```toml title="~/.config/chezmoi/chezmoi.toml"
[diff]
    pager = "diff-so-fancy"
```

The pager can be disabled using the `--no-pager` flag or by setting `diff.pager`
to an empty string.

## Show human-friendly diffs for binary files

Similar to git, chezmoi includes a "textconv" feature that can transform file
contents before passing them to the diff program. This is primarily useful for
generating human-readable diffs of binary files.

For example, to show diffs of macOS `.plist` files, add the following to your
configuration file:

=== "JSON"

    ```json title="~/.config/chezmoi/chezmoi.json"
    {
        "textconv": [
            "pattern": "**/*.plist",
            "command": "plutil",
            "args": [
                "-convert",
                "xml1",
                "-o",
                "-",
                "-"
            ]
        ]
    }
    ```

=== "TOML"

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [[textconv]]
    pattern = "**/*.plist"
    command = "plutil"
    args = ["-convert", "xml1", "-o", "-", "-"]
    ```

=== "YAML"

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    textconv:
    - pattern: "**/*.plist"
      command: "plutil"
      args:
      - "-convert"
      - "xml1"
      - "-o"
      - "-",
      - "-"
    ```

This will pipe all `.plist` files through `plutil -convert xml1 -o - -` before
showing differences.

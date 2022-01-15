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

## Don't show scripts in the diff output

By default, `chezmoi diff` will show all changes, including the contents of
scripts that will be run. You can exclude scripts from the diff output by
setting the `diff.exclude` configuration variable in your configuration file,
for example:

```toml title="~/.config/chezmoi/chezmoi.toml"
[diff]
    exclude = ["scripts"]
```

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

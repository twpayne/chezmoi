# Manage machine-to-machine differences

## Use templates

The primary goal of chezmoi is to manage configuration files across multiple
machines, for example your personal macOS laptop, your work Ubuntu desktop, and
your work Linux laptop. You will want to keep much configuration the same
across these, but also need machine-specific configurations for email
addresses, credentials, etc. chezmoi achieves this functionality by using
[`text/template`](https://pkg.go.dev/text/template) for the source state where
needed.

For example, your home `~/.gitconfig` on your personal machine might look like:

```toml title="~/.gitconfig"
[user]
    email = "me@home.org"
```

Whereas at work it might be:

```toml title="~/.gitconfig"
[user]
    email = "firstname.lastname@company.com"
```

To handle this, on each machine create a configuration file called
`~/.config/chezmoi/chezmoi.toml` defining variables that might vary from
machine to machine. For example, for your home machine:

```toml title="~/.config/chezmoi/chezmoi.toml"
[data]
    email = "me@home.org"
```

If you intend to store private data (e.g. access tokens) in
`~/.config/chezmoi/chezmoi.toml`, make sure it has permissions `0600`.

If you prefer, you can use JSON, JSONC, or YAML for your configuration file.
Variable names must start with a letter and be followed by zero or more letters
or digits.

Then, add `~/.gitconfig` to chezmoi using the `--template` flag to turn it
into a template:

```console
$ chezmoi add --template ~/.gitconfig
```

You can then open the template (which will be saved in the file
`~/.local/share/chezmoi/dot_gitconfig.tmpl`):

```console
$ chezmoi edit ~/.gitconfig
```

Edit the file so it looks something like:

```toml title="~/.local/share/chezmoi/dot_gitconfig.tmpl"
[user]
    email = {{ .email | quote }}
```

Templates are often used to capture machine-specific differences. For example,
in your `~/.local/share/chezmoi/dot_bashrc.tmpl` you might have:

``` title="~/.local/share/chezmoi/dot_bashrc.tmpl"
# common config
export EDITOR=vi

# machine-specific configuration
{{- if eq .chezmoi.hostname "work-laptop" }}
# this will only be included in ~/.bashrc on work-laptop
{{- end }}
```

For a full list of variables, run:

```console
$ chezmoi data
```

For more advanced usage, you can use the full power of the
[`text/template`](https://pkg.go.dev/text/template) language. chezmoi includes
all of the text functions from [sprig](http://masterminds.github.io/sprig/) and
its own [functions for interacting with password
managers](../reference/templates/functions/index.md).

Templates can be executed directly from the command line, without the need to
create a file on disk, with the `execute-template` command, for example:

```console
$ chezmoi execute-template "{{ .chezmoi.os }}/{{ .chezmoi.arch }}"
```

This is useful when developing or debugging templates.

Some password managers allow you to store complete files. The files can be
retrieved with chezmoi's template functions. For example, if you have a file
stored in 1Password with the UUID `uuid` then you can retrieve it with the
template:

```
{{- onepasswordDocument "uuid" -}}
```

The `-`s inside the brackets remove any whitespace before or after the template
expression, which is useful if your editor has added any newlines.

If, after executing the template, the file contents are empty, the target file
will be removed. This can be used to ensure that files are only present on
certain machines. If you want an empty file to be created anyway, you will need
to give it an `empty_` prefix.

## Ignore files or a directory on different machines

For coarser-grained control of files and entire directories managed on
different machines, or to exclude certain files completely, you can create
`.chezmoiignore` files in the source directory. These specify a list of
patterns that chezmoi should ignore, and are interpreted as templates. An
example `.chezmoiignore` file might look like:

``` title="~/.local/share/chezmoi/.chezmoiignore"
README.md
{{- if ne .chezmoi.hostname "work-laptop" }}
.work # only manage .work on work-laptop
{{- end }}
```

The use of `ne` (not equal) is deliberate. What we want to achieve is "only
install `.work` if hostname is `work-laptop`" but chezmoi installs everything
by default, so we have to turn the logic around and instead write "ignore
`.work` unless the hostname is `work-laptop`".

Patterns can be excluded by starting the line with a `!`, for example:

``` title="~/.local/share/chezmoi/.chezmoiignore"
dir/f*
!dir/foo
```

will ignore all files beginning with an `f` in `dir` except for `dir/foo`.

You can see what files chezmoi ignores with the command

```console
$ chezmoi ignored
```

## Handle different file locations on different systems with the same contents

If you want to have the same file contents in different locations on different
systems, but maintain only a single file in your source state, you can use a
shared template.

Create the common file in the `.chezmoitemplates` directory in the source
state. For example, create `.chezmoitemplates/file.conf`. The contents of this
file are available in templates with the `template $NAME .` function where
`$NAME` is the name of the file (`.` passes the current data to the template
code in `file.conf`; see 
[`template` action](https://pkg.go.dev/text/template#hdr-Actions) for details).

Then create files for each system, for example `Library/Application
Support/App/file.conf.tmpl` for macOS and `dot_config/app/file.conf.tmpl` for
Linux. Both template files should contain `{{- template "file.conf" . -}}`.

Finally, tell chezmoi to ignore files where they are not needed by adding lines
to your `.chezmoiignore` file, for example:

``` title="~/.local/share/chezmoi/.chezmoiignore"
{{ if ne .chezmoi.os "darwin" }}
Library/Application Support/App/file.conf
{{ end }}
{{ if ne .chezmoi.os "linux" }}
.config/app/file.conf
{{ end }}
```

## Use completely different dotfiles on different machines

chezmoi's template functionality allows you to change a file's contents based
on any variable. For example, if you want `~/.bashrc` to be different on Linux
and macOS you would create a file in the source state called `dot_bashrc.tmpl`
containing:

``` title="~/.local/share/chezmoi/dot_bashrc.tmpl"
{{ if eq .chezmoi.os "darwin" -}}
# macOS .bashrc contents
{{ else if eq .chezmoi.os "linux" -}}
# Linux .bashrc contents
{{ end -}}
```

However, if the differences between the two versions are so large that you'd
prefer to use completely separate files in the source state, you can achieve
this with the `include` template function.

Create the following files:

```bash title="~/.local/share/chezmoi/.bashrc_darwin"
# macOS .bashrc contents
```

```bash title="~/.local/share/chezmoi/.bashrc_linux"
# Linux .bashrc contents
```

``` title="~/.local/share/chezmoi/dot_bashrc.tmpl"
{{- if eq .chezmoi.os "darwin" -}}
{{-   include ".bashrc_darwin" -}}
{{- else if eq .chezmoi.os "linux" -}}
{{-   include ".bashrc_linux" -}}
{{- end -}}
```

This will cause `~/.bashrc` to contain `~/.local/share/chezmoi/.bashrc_darwin`
on macOS and `~/.local/share/chezmoi/.bashrc_linux` on Linux.

If you want to use templates within your templates, then, instead, create:

```bash title="~/.local/share/chezmoi/.chezmoitemplates/bashrc_darwin.tmpl"
# macOS .bashrc template contents
```

```bash title="~/.local/share/chezmoi/.chezmoitemplates/bashrc_linux.tmpl"
# Linux .bashrc template contents
```

``` title="~/.local/share/chezmoi/dot_bashrc.tmpl"
{{- if eq .chezmoi.os "darwin" -}}
{{-   template "bashrc_darwin.tmpl" . -}}
{{- else if eq .chezmoi.os "linux" -}}
{{-   template "bashrc_linux.tmpl" . -}}
{{- end -}}
```

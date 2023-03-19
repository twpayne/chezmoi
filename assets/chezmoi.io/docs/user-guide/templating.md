# Templating

## Introduction

Templates are used to change the contents of a file depending on the
environment. For example, you can use the hostname of the machine to create
different configurations on different machines.

chezmoi uses the [`text/template`](https://pkg.go.dev/text/template) syntax from
Go extended with [text template functions from
`sprig`](http://masterminds.github.io/sprig/).

When reading files from the source state, chezmoi interprets them as a template
if either of the following is true:

* The file name has a `.tmpl` suffix.

* The file is in the `.chezmoitemplates` directory, or a subdirectory of
  `.chezmoitemplates`.

## Template data

chezmoi provides a variety of template variables. For a full list, run

```console
$ chezmoi data
```

These come from a variety of sources (later data overwrite earlier ones):

* Variables populated by chezmoi are in `.chezmoi`, for example `.chezmoi.os`.

* Variables created by you in the `.chezmoidata.$FORMAT` configuration file.
  The various supported formats (`json`, `jsonc`, `toml` and `yaml`) are read in
  alphabetical order.

* Variables created by you in the `data` section of the configuration file.

Furthermore, chezmoi provides a variety of functions to retrieve data at runtime
from password managers, environment variables, and the filesystem.

## Creating a template file

There are several ways to create a template:

* When adding a file for the first time, pass the `--template` argument, for example:

    ```console
    $ chezmoi add --template ~/.zshrc
    ```

* If a file is already managed by chezmoi, but is not a template, you can make
  it a template by running, for example:

    ```console
    $ chezmoi chattr +template ~/.zshrc
    ```

* You can create a template manually in the source directory by giving it a
  `.tmpl` extension, for example:

    ```console
    $ chezmoi cd
    $ $EDITOR dot_zshrc.tmpl
    ```

* Templates in `.chezmoitemplates` must be created manually, for example:

    ```console
    $ chezmoi cd
    $ mkdir -p .chezmoitemplates
    $ cd .chezmoitemplates
    $ $EDITOR mytemplate
    ```

## Editing a template file

The easiest way to edit a template is to use `chezmoi edit`, for example:

```console
$ chezmoi edit ~/.zshrc
```

This will open the source file for `~/.zshrc` in `$EDITOR`. When you quit the
editor, chezmoi will check the template syntax.

If you want the changes you make to be immediately applied after you quit the
editor, use the `--apply` option, for example:

```console
$ chezmoi edit --apply ~/.zshrc
```

## Testing templates

Templates can be tested with the `chezmoi execute-template` command which treats
each of its arguments as a template and executes it. This can be useful for
testing small fragments of templates, for example:

```console
$ chezmoi execute-template '{{ .chezmoi.hostname }}'
```

If there are no arguments, `chezmoi execute-template` will read the template
from the standard input. This can be useful for testing whole files, for example:

```console
$ chezmoi cd
$ chezmoi execute-template < dot_zshrc.tmpl
```

## Template syntax

Template actions are written inside double curly brackets, `{{` and `}}`.
Actions can be variables, pipelines, or control statements. Text outside actions
is copied literally.

Variables are written literally, for example:

```
{{ .chezmoi.hostname }}
```

Conditional expressions can be written using `if`, `else if`, `else`, and `end`,
for example:

```
{{ if eq .chezmoi.os "darwin" }}
# darwin
{{ else if eq .chezmoi.os "linux" }}
# linux
{{ else }}
# other operating system
{{ end }}
```

For a full description of the template syntax, see the [`text/template`
documentation](https://pkg.go.dev/text/template).

### Removing whitespace

For formatting reasons you might want to leave some whitespace after or before
the template code. This whitespace will remain in the final file, which you
might not want.

A solution for this is to place a minus sign and a space next to the brackets.
So `{{- ` for the left brackets and ` -}}` for the right brackets. Here's an
example:

```
HOSTNAME={{- .chezmoi.hostname }}
```

This will result in

```
HOSTNAME=myhostname
```

Notice that this will remove any number of tabs, spaces and even newlines and
carriage returns.

## Debugging templates

If there is a mistake in one of your templates and you want to debug it, chezmoi
can help you. You can use this subcommand to test and play with the examples in
these docs as well.

There is a very handy subcommand called `execute-template`. chezmoi will
interpret any data coming from stdin or at the end of the command. It will then
interpret all templates and output the result to stdout. For example with the
command:

```console
$ chezmoi execute-template '{{ .chezmoi.os }}/{{ .chezmoi.arch }}'
```

chezmoi will output the current OS and architecture to stdout.

You can also feed the contents of a file to this command by typing:

```console
$ cat foo.txt | chezmoi execute-template
```

## Simple logic

A very useful feature of chezmoi templates is the ability to perform logical
operations.

```
# common config
export EDITOR=vi

# machine-specific configuration
{{- if eq .chezmoi.hostname "work-laptop" }}
# this will only be included in ~/.bashrc on work-laptop
{{- end }}
```

In this example chezmoi will look at the hostname of the machine and if that is
equal to "work-laptop", the text between the `if` and the `end` will be included
in the result.

### Boolean functions

| Function | Return value                                                                                                                                                                              |
| -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `eq`     | Returns true if the first argument is equal to any of the other arguments                                                                                                                 |
| `not`    | Returns the boolean negation of its single argument                                                                                                                                       |
| `and`    | Returns the boolean AND of its arguments by returning the first empty argument or the last argument, that is, `and x y` behaves as `if x then y else x`. All the arguments are evaluated  |
| `or`     | Returns the boolean OR of its arguments by returning the first non-empty argument or the last argument, that is, `or x y` behaves as `if x then x else y` All the arguments are evaluated |

### Integer functions

| Function | Return value                               |
| -------- | ------------------------------------------ |
| `len`    | Returns the integer length of its argument |
| `eq`     | Returns the boolean truth of arg1 == arg2  |
| `ne`     | Returns the boolean truth of arg1 != arg2  |
| `lt`     | Returns the boolean truth of arg1 < arg2   |
| `le`     | Returns the boolean truth of arg1 <= arg2  |
| `gt`     | Returns the boolean truth of arg1 > arg2   |
| `ge`     | Returns the boolean truth of arg1 >= arg2  |

## More complicated logic

Up until now, we have only seen if statements that can handle at most two
variables. In this part we will see how to create more complicated expressions.

You can also create more complicated expressions. The `eq` command can accept
multiple arguments. It will check if the first argument is equal to any of the
other arguments.

```
{{ if eq "foo" "foo" "bar" }}hello{{end}}
{{ if eq "foo" "bar" "foo" }}hello{{end}}
{{ if eq "foo" "bar" "bar" }}hello{{end}}
```

The first two examples will output `hello` and the last example will output
nothing.

The operators `or` and `and` can also accept multiple arguments.

### Chaining operators

You can perform multiple checks in one if statement.

```
{{ if (and (eq .chezmoi.os "linux") (ne .email "me@home.org")) }}
...
{{ end }}
```

This will check if the operating system is Linux and the configured email is not
the home email. The brackets are needed here, because otherwise all the
arguments will be give to the `and` command.

This way you can chain as many operators together as you like.

## Helper functions

chezmoi has added multiple helper functions to the
[`text/template`](https://pkg.go.dev/text/template) syntax.

chezmoi includes [`sprig`](http://masterminds.github.io/sprig/), an extension to
the `text/template` format that contains many helper functions. Take a look at
their documentation for a list.

chezmoi adds a few functions of its own as well. Take a look at the
[reference](/reference/templates/functions/) for complete list.

## Template variables

chezmoi defines a few useful templates variables that depend on the system you
are currently on. A list of the variables defined by chezmoi can be found
[here](/reference/templates/variables/).

There are, however more variables than that. To view the variables available on
your system, execute:

```console
$ chezmoi data
```

This outputs the variables in JSON format by default. To access the variable
`chezmoi.kernel.osrelease` in a template, use

```
{{ .chezmoi.kernel.osrelease }}
```

This way you can also access the variables you defined yourself.

## Using `.chezmoitemplates`

Files in the `.chezmoitemplates` subdirectory are parsed as templates and are
available to be included in other templates using the [`template`
action](https://pkg.go.dev/text/template#hdr-Actions) with a name equal to their
relative path to the `.chezmoitemplates` directory.

By default, such templates will be executed with `nil` data. If you want to
access template variables (e.g. `.chezmoi.os`) in the template you must pass the
data explicitly.

For example:

```
.chezmoitemplates/part.tmpl:
{{ if eq .chezmoi.os "linux" }}
# linux config
{{ else }}
# non-linux config
{{ end }}

dot_file.tmpl:
{{ template "part.tmpl" . }}
```

## Using `.chezmoitemplates` for creating similar files

When you have multiple similar files, but they aren't quite the same, you can
create a template file in the directory `.chezmoitemplates`. This template can
be inserted in other template files, for example:

Create `.local/share/chezmoi/.chezmoitemplates/alacritty`:

```
some: config
fontsize: {{ . }}
more: config
```

Notice the file name doesn't have to end in `.tmpl`, as all files in the
directory `.chezmoitemplates` are interpreted as templates.

Create other files using the template
`~/.local/share/chezmoi/small-font.yml.tmpl`

```
{{- template "alacritty" 12 -}}
```

`~/.local/share/chezmoi/big-font.yml.tmpl`

```
{{- template "alacritty" 18 -}}
```

Here we're calling the shared `alacritty` template with the font size as the
`.` value passed in. You can test this with `chezmoi cat`:

```console
$ chezmoi cat ~/small-font.yml
some: config
fontsize: 12
more: config
$ chezmoi cat ~/big-font.yml
some: config
fontsize: 18
more: config
```

### Passing multiple arguments

In the example above only one arguments is passed to the template. To pass more
arguments to the template, you can do it in two ways.

#### Via the config file

This method is useful if you want to use the same template arguments multiple
times, because you don't specify the arguments every time. Instead you specify
them in the file `~/.config/chezmoi/chezmoi.toml`:

```toml title="~/.config/chezmoi/chezmoi.toml"
[data.alacritty.big]
    fontsize = 18
    font = "DejaVu Serif"
[data.alacritty.small]
    fontsize = 12
    font = "DejaVu Sans Mono"
```

Use the variables in `~/.local/share/chezmoi/.chezmoitemplates/alacritty`:

``` title="~/.local/share/chezmoi/.chezmoitemplates/alacritty"
some: config
fontsize: {{ .fontsize }}
font: {{ .font }}
more: config
```

And connect them with `~/.local/share/chezmoi/small-font.yml.tmpl`:

``` title="~/.local/share/chezmoi/small-font.yml.tmpl"
{{- template "alacritty" .alacritty.small -}}
```

At the moment, this means that you'll have to duplicate the alacritty data in
the config file on every machine, but a feature will be added to avoid this.

#### By passing a dictionary

Using the same alacritty configuration as above, you can pass the arguments to
it with a dictionary, for example `~/.local/share/chezmoi/small-font.yml.tmpl`:

``` title="~/.local/share/chezmoi/small-font.yml.tmpl"
{{- template "alacritty" dict "fontsize" 12 "font" "DejaVu Sans Mono" -}}
```

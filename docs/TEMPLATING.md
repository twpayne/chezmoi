# chezmoi Templating Guide

<!--- toc ---> 
* [Introduction](#introduction)
* [Creating a template file](#creating-a-template-file)
* [Debugging templates](#debugging-templates)
* [Simple logic](#simple-logic)
* [More complicated logic](#more-complicated-logic)
* [Helper functions](#helper-functions)
* [Template variables](#template-variables)
* [Using .chezmoitemplates for creating similar files](#using-chezmoitemplates-for-creating-similar-files)

## Introduction

Templates are used to create different configurations depending on the enviorment.
For example, you can use the hostname of the machine to create different
configurations.

chezmoi uses the [`text/template`](https://pkg.go.dev/text/template) syntax from
Go, extended with [text template functions from `sprig`](http://masterminds.github.io/sprig/)
You can look there for more information.

## Creating a template file

chezmoi will not interpret all files as templates. It will only do that if the 
filename ends with .tmpl or it is in the .chezmoitemplates directory.

There are a few ways to create a template file in chezmoi. 
If the file is not yet known by chezmoi you can do the following:

	chezmoi add ~/.zshrc --template

This will add ~/.zshrc as a template to the source state. This means that chezmoi
will add a .tmpl extension to file and interpret any templates in the source upon
updating.

You can also use the command

	chezmoi add ~/.zshrc --autotemplate

to add ~/.zshrc to the source state as a template, while replacing any strings
that it can match with the variables from the data section of the chezmoi config.

If the file is already known by chezmoi, you can use the command

	chezmoi chattr template ~/.zshrc

Or you can simply add the file extension .tmpl to the file in the source directory.
This way chezmoi will interpret the file as a template.

## Template syntax

Every template expression starts and ends with double curly brackets ('{{' and '}}').
Between these brackets can be either variables or functions.

An example with a variable

	{{.chezmoi.hostname}}

An example with a function

	{{if expression}} Some text {{end}}

If the result of the expression is empty (false, 0, empty string, ...), no output
will be generated. Otherwise this will result in the text in between the if and the 
end.

### Remove whitespace

For formatting reasons you might want to leave some whitespace after or before the 
template code. This whitespace will remain in the final file, which you might not want.

A solution for this is to place a minus sign and a space next to the brackets. So
'{{- ' for the left brackets and ' -}}' for the right brackets. Here's an example:

	HOSTNAME=  			{{- .chezmoi.hostname }}

This will result in

	HOSTNAME=myhostname

Notice that this will remove any number of tabs, spaces and even newlines and carriage
returns.

## Debugging templates

If there is a mistake in one of your templates and you want to debug it, chezmoi
can help you. You can use this subcommand to test and play with the examples in these
docs as well.

There is a very handy subcommand called "execute-template". chezmoi will interpret
any data coming from stdin or at the end of the command. It will then interpret all
templates and output the result to stdout.
For example with the command:

	chezmoi execute-template '{{ .chezmoi.os }}/{{ .chezmoi.arch }}'

chezmoi will output the current os and architecture to stdout.

You can also feed the contents of a file to this command by typing:

	cat foo.txt | chezmoi exectute-template

## Simple logic

A very useful feature of chezmoi templates is the ability to perform logical operations.

	# common config
	export EDITOR=vi
	
	# machine-specific configuration
	{{- if eq .chezmoi.hostname "work-laptop" }}
	# this will only be included in ~/.bashrc on work-laptop
	{{- end }}

In this example chezmoi will look at the hostname of the machine and if that is equal to
"work-laptop", the text between the "if" and the "end" will be included in the result.

### Locical operators

The following operators are available:

* `eq`  - Return true if the first argument is equal to any other argument.
* `or`  - Return boolean or of the arguments.
* `and` - Return boolean and of the arguments.
* `not` - Return boolean negative of the argument.
* `len` - Return the length of the argument.

Notice that some operators can accept more than two arguments.

### Integer operators

There are separate operators for comparing integers.

* `eq` - Return true if the first argument is equal to any other argument. - arg1 == arg2 		 
* `ne` - Returns if arg1 is not equal to arg2                              - arg1 != arg2
* `lt` - Returns if arg1 is less than arg2.                                - arg1 <  arg2 
* `le` - Returns if arg1 is less than or equal to arg2.                    - arg1 <= arg2
* `gt` - Returns if arg1 is greater than arg2.                             - arg1 >  arg2
* `ge` - Returns if arg1 is greater than or equal to arg2.                 - arg1 >= arg2

`eq` can handle multiple arguments again, the same way as the "eq" above.

## More complicated logic

Up until now, we have only seen if statements that can handle at most two variables.
In this part we will see how to create more complicated expressions.

You can also create more complicated expressions. The `eq` command can accept multiple
arguments. It will check if the first argument is equal to any of the other arguments.
	
	{{ if eq "foo" "foo" "bar" }}hello{{end}}

	{{ if eq "foo" "bar" "foo" }}hello{{end}}

	{{ if eq "foo" "bar" "bar" }}hello{{end}}

The first two examples will output "hello" and the last example will output nothing.

The operators `or` and `and` can also accept multiple arguments.

### Chaining operators

You can perform multiple checks in one if statement.

	{{ and ( eq .chezmoi.os "linux" ) ( ne .email "john@home.org" ) }}

This will check if the operating system is Linux and the configured email
is not the home email. The brackets are needed here, because otherwise all the
arguments will be give to the `and` command.

This way you can chain as many operators together as you like.

## Helper functions

chezmoi has added multiple helper functions to the [`text/template`](https://pkg.go.dev/text/template) 
syntax.  

Chezmoi includes [`Sprig`](http://masterminds.github.io/sprig/), an extension to 
the text/template format that contains many helper functions. Take a look at 
their documentation for a list.

Chezmoi adds a few functions of its own as well. Take a look at the 
[`reference`](REFERENCE.md#template-functions) for complete list.

## Template variables

Chezmoi defines a few useful templates variables that depend on the system
you are currently on. A list of the variables defined by chezmoi can be found 
[here](REFERENCE.md#template-variables).

There are, however more variables than
that. To view the variables available on your system, execute:

	chezmoi data

This outputs the variables in JSON format by default. To access the variable
`chezmoi>kernel>osrelease` in a template, use

	{{ .chezmoi.kernel.osrelease }}

This way you can also access the variables you defined yourself.

## Using .chezmoitemplates for creating similar files

When you have multiple similar files, but they aren't quite the same, you can create
a template file in the directory .chezmoitemplates. This template can be inserted
in other template files. 
For example:

Create:

	.local/share/chezmoi/.chezmoitemplates/alacritty:

Notice the file name doesn't have to end in .tmpl, as all files in the directory
.chemzoitemplates are interpreted as templates.

	some: config
	fontsize: {{ . }}
	somemore: config

Create other files using the template

`.local/share/chezmoi/small-font.yml.tmpl`

```
{{- template "alacritty" 12 -}}
```

`.local/share/chezmoi/big-font.yml.tmpl`

```
{{- template "alacritty" 18 -}}
```

Here we're calling the shared `alacritty` template with the he font size as 
the `.` value passed in. You can test this with `chezmoi cat`:

	$ chezmoi cat ~/small-font.yml
	some: config
	fontsize: 12
	somemore: config
	$ chezmoi cat ~/big-font.yml
	some: config
	fontsize: 18
	somemore: config

### Passing multiple arguments
In the example above only one arguments is passed to the template. To pass
more arguments to the template, you can do it in two ways.

#### Via chezmoi.toml

This method is useful if you want to use the same template arguments multiple
times, because you don't specify the arguments every time. Instead you specify
them in the file `.chezmoi.toml`.

`.config/chezmoi/chezmoi.toml`:

```
[data.alacritty.big]
  fontsize = 18
  font = DejaVu Serif
[data.alacritty.small]
  fontsize = 12
  font = DejaVu Sans Mono
```

`.local/share/chezmoi/.chezmoitemplates/alacritty`:

```
some: config
fontsize: {{ .fontsize }}
font: {{ .font }}
somemore: config
```

`.local/share/chezmoi/small-font.yml.tmpl`

```
{{- template "alacritty" .alacritty.small -}}
```

`.local/share/chezmoi/big-font.yml.tmpl`

```
{{- template "alacritty" .alacritty.big -}}
```

At the moment, this means that you'll have to duplicate the alacritty data in 
the config file on every machine, but a feature will be added to avoid this.

#### By passing a dictionary

Using the same alacritty configuration as above, you can pass the arguments to
it with a dictionary.

`.local/share/chezmoi/small-font.yml.tmpl`

```
{{- template "alacritty" dict "fontsize" 12 "font" "DejaVu Sans Mono" -}}
```

`.local/share/chezmoi/big-font.yml.tmpl`

```
{{- template "alacritty" dict "fontsize" 18 "font" "DejaVu Serif" -}}
```

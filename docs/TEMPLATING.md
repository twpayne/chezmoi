# chezmoi Templating Guide

<!--- toc ---> 
* [Introduction](#introduction)
* [Creating a template file](#creating-a-template-file)
* [Debugging templates](#debugging-templates)
* [Simple logic](#simple-logic)
* [Using .chezmoitemplates for creating similar files](#using-chezmoitemplates-for-creating-similar-files)

## Introduction

Templates are used to create different configurations depending on the enviorment.
For example, you can use the hostname of the machine to create different
configurations.

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

If the file is already known by chezmoi, you can simply add the file extension
.tmpl to the file in the source directory. This way chezmoi will interpret the file
as a template.

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

In this example chezmoi will look at the hostname of the machine and change the
contents of the resulting file based on that information.

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

This approach only works for a single value. If you want to pass in more than one value you can pass in structured data from your config file:

`.config/chezmoi/chezmoi.toml`

```
[data.alacritty.big]
  fontsize = 18
[data.alacritty.small]
  fontsize = 12
```

`.local/share/chezmoi/.chezmoitemplates/alacritty`:

```
some: config
fontsize: {{ .fontsize }}
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

At the moment, this means that you'll have to duplicate the alacritty data in the config file on every machine, 
but a feature will be added to avoid this.

# `execute-template` [*template*...]

Execute *template*s. This is useful for testing templates or for calling
chezmoi from other scripts. *templates* are interpreted as literal templates,
with no whitespace added to the output between arguments. If no templates are
specified, the template is read from stdin.

## `--init`, `-i`

Include simulated functions only available during `chezmoi init`.

## `--promptBool` *pairs*

Simulate the `promptBool` template function with a function that returns values
from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptBool` is called with a *prompt* that does not match any of *pairs*, then
it returns false.

## `--promptInt` *pairs*

Simulate the `promptInt` template function with a function that returns values
from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptInt` is called with a *prompt* that does not match any of *pairs*, then
it returns zero.

## `--promptString`, `-p` *pairs*

Simulate the `promptString` template function with a function that returns
values from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value*
pairs. If `promptString` is called with a *prompt* that does not match any of
*pairs*, then it returns *prompt* unchanged.

## `--stdinisatty` *bool*

Simulate the `stdinIsATTY` function by returning *bool*.

!!! example

    ```console
    $ chezmoi execute-template '{{ .chezmoi.sourceDir }}'
    $ chezmoi execute-template '{{ .chezmoi.os }}' / '{{ .chezmoi.arch }}'
    $ echo '{{ .chezmoi | toJson }}' | chezmoi execute-template
    $ chezmoi execute-template --init --promptString email=me@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl
    ```

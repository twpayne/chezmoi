# `execute-template` [*template*...]

Execute *template*s. This is useful for [testing
templates](../../user-guide/templating.md#testing-templates) or for calling
chezmoi from other scripts. *templates* are interpreted as literal templates,
with no whitespace added to the output between arguments. If no templates are
specified, the template is read from stdin.

## Flags

### `-i`, `--init`

Include simulated functions only available during `chezmoi init`.

### `--left-delimiter` *delimiter*

Set the left template delimiter.

### `--promptBool` *pairs*

Simulate the `promptBool` template function with a function that returns values
from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptBool` is called with a *prompt* that does not match any of *pairs*, then
it returns false.

### `--promptChoice` *pairs*

Simulate the `promptChoice` template function with a function that returns
values from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value*
pairs. If `promptChoice` is called with a *prompt* that does not match any of
*pairs*, then it returns false.

### `--promptInt` *pairs*

Simulate the `promptInt` template function with a function that returns values
from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value* pairs. If
`promptInt` is called with a *prompt* that does not match any of *pairs*, then
it returns zero.

### `-p`, `--promptString` *pairs*

Simulate the `promptString` template function with a function that returns
values from *pairs*. *pairs* is a comma-separated list of *prompt*`=`*value*
pairs. If `promptString` is called with a *prompt* that does not match any of
*pairs*, then it returns *prompt* unchanged.

### `--right-delimiter` *delimiter*

Set the right template delimiter.

### `--stdinisatty` *bool*

Simulate the `stdinIsATTY` function by returning *bool*.

### `--with-stdin`

If run with arguments, then set `.chezmoi.stdin` to the contents of the standard
input.

## Examples

```console
$ chezmoi execute-template '{{ .chezmoi.sourceDir }}'
$ chezmoi execute-template '{{ .chezmoi.os }}' / '{{ .chezmoi.arch }}'
$ echo '{{ .chezmoi | toJson }}' | chezmoi execute-template
$ chezmoi execute-template --init --promptString email=me@home.org < ~/.local/share/chezmoi/.chezmoi.toml.tmpl
```

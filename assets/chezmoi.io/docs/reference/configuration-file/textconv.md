# textconv

A section called `textconv` in the configuration file controls how file contents
are modified before being passed to diff.

The `textconv` must contain an array of objects where each object has the
following properties:

| Name      | Type     | Description                   |
| --------- | -------- | ----------------------------- |
| `pattern` | string   | Target path pattern to match  |
| `command` | string   | Command to transform contents |
| `args`    | []string | Extra arguments to command    |

Files whose target path matches `pattern` are transformed by passing them to the
standard input of `command` with `args`, and new contents are read from the
command's standard output.

If a target path does not match any patterns then the file contents are passed
unchanged to diff. If a target path matches multiple patterns then element with
the longest `pattern` is used.

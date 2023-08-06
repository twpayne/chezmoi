# HuJSON - "Human JSON" ([JWCC](https://nigeltao.github.io/blog/2021/json-with-commas-comments.html))

The `github.com/tailscale/hujson` package implements
the [JWCC](https://nigeltao.github.io/blog/2021/json-with-commas-comments.html) extension
of [standard JSON](https://datatracker.ietf.org/doc/html/rfc8259).

The `JWCC` format permits two things over standard JSON:

1. C-style line comments and block comments intermixed with whitespace,
2. allows trailing commas after the last member/element in an object/array.

All JSON is valid JWCC.

For details, see the JWCC docs at:

https://nigeltao.github.io/blog/2021/json-with-commas-comments.html

## Visual Studio Code association

Visual Studio Code supports a similar `jsonc` (JSON with comments) format. To
treat all `*.hujson` files as `jsonc` with trailing commas allowed, you can add
the following snippet to your Visual Studio Code configuration:

```json
"files.associations": {
    "*.hujson": "jsonc"
},
"json.schemas": [{
    "fileMatch": ["*.hujson"],
    "schema": {
        "allowTrailingCommas": true
    }
}]
```
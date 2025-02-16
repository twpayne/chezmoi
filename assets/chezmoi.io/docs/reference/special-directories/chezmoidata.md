# `.chezmoidata/`

If any `.chezmoidata/` directories exist in the source state, all files within
them are interpreted as structured static data in the given formats. This data
can then be used in templates. See also [`.chezmoidata.$FORMAT`][data-format].

--8<-- "config-format.md"

!!! info

    The files in the `.chezmoidata` directories all *merge* to the root of the
    data dictionary and are read in lexical (alphabetic) filesystem order. This
    applies both within `.chezmoidata` directories and between `.chezmoidata`
    directories.

    As an example, if I have a `.chezmoidata` directory in my `dot_config`
    source directory, the files within will be merged according to the sort
    order of the files:

    === "JSON"

        ```json title="dot_config/.chezmoidata/zed.json"
        { "z": { "z": 3 } }
        ```

    === "JSONC"

        ```jsonc title="dot_config/.chezmoidata/alpha.jsonc"
        { "z": { "z": 4 } }
        ```

    === "TOML"

        ```toml title="dot_config/.chezmoidata/beta.toml"
        z.x = 1
        ```

    === "YAML"

        ```toml title="dot_config/.chezmoidata/gamma.yaml"
        z:
          y: 2
        ```

    The output of `chezmoi data` will include the following merged `z`
    dictionary. Note that the value in `.chezmoidata/zed.json` overwrote the
    value in `.chezmoidata/alpha.jsonc` because of the lexical file sorting.

    ```json
    {
      "z": {
        "x": 1,
        "y": 2,
        "z": 3
      }
    }
    ```

!!! warning

    Files in `.chezmoidata` directories cannot be templates because they must be
    present prior to the start of the template engine. Dynamic machine data
    should be set in the `data` section of [`.chezmoi.$FORMAT.tmpl`][config].
    Dynamic environment data should be read from templates using the
    [`output`][output], [`fromJson`][fromjson], [`fromJson`][fromjson], or
    similar functions.

[data-format]: /reference/special-files/chezmoidata-format.md
[config]: /reference/special-files/chezmoidata-format.md
[fromjson]: /reference/templates/functions/fromJson.md
[fromyaml]: /reference/templates/functions/fromYaml.md
[output]: /reference/templates/functions/output.md

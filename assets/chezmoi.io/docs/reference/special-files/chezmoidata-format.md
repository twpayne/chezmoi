# `.chezmoidata.$FORMAT`

If `.chezmoidata.$FORMAT` files exist in the source state, they are interpreted
as structured static data in the given format. This data can then be used in
templates. See also [`.chezmoidata/`][chezmoidata-dir].

!!! example

    If `.chezmoidata.toml` contains the following:

    ```toml title="~/.local/share/chezmoi/.chezmoidata.toml"
    fontSize = 12
    ```

    Then the `.fontSize` variable is available in templates, e.g.

    ```
    FONT_SIZE={{ .fontSize }}
    ```

    Will result in:

    ```
    FONT_SIZE=12
    ```

--8<-- "config-format.md"

!!! info

    There may be multiple `.chezmoidata.$FORMAT` files in the source state. They
    all *merge* to the root of the data dictionary and they are read in lexical
    (alphabetic) filesystem order.

    As an example, if I have four `.chezmoidata.$FORMAT` files in my
    `dot_config` source directory, they will be merged according to the sort
    order of the files:

    === "JSON"

        ```json title="dot_config/.chezmoidata.json"
        { "z": { "z": 3 } }
        ```

    === "JSONC"

        ```jsonc title="dot_config/.chezmoidata.jsonc"
        { "z": { "z": 4 } }
        ```

    === "TOML"

        ```toml title="dot_config/.chezmoidata.toml"
        z.x = 1
        ```

    === "YAML"

        ```toml title="dot_config/.chezmoidata.yaml"
        z:
          y: 2
        ```

    The output of `chezmoi data` will include the following merged `z`
    dictionary. Note that the value in `.chezmoidata.jsonc` overwrote the value
    in `.chezmoidata.json` because of the lexical file sorting.

    ```json
    {
      "z": {
        "x": 1,
        "y": 2,
        "z": 4
      }
    }
    ```

!!! warning

    `.chezmoidata.$FORMAT` files cannot be templates because they must be
    present prior to the start of the template engine. Dynamic machine data
    should be set in the `data` section of [`.chezmoi.$FORMAT.tmpl`][config].
    Dynamic environment data should be read from templates using the
    [`output`][output], [`fromJson`][fromjson], [`fromJson`][fromjson], or
    similar functions.

[config]: /reference/special-files/chezmoidata-format.md
[fromjson]: /reference/templates/functions/fromJson.md
[fromyaml]: /reference/templates/functions/fromYaml.md
[output]: /reference/templates/functions/output.md
[chezmoidata-dir]: /reference/special-directories/chezmoidata.md

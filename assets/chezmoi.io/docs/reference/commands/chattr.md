# `chattr` *modifier* *target*...

Change the attributes and/or type of *target*s. *modifier* specifies what to
modify.

Add attributes by specifying them or their abbreviations directly, optionally
prefixed with a plus sign (`+`). Remove attributes by prefixing them or their
attributes with the string `no` or a minus sign (`-`). The available attribute
modifiers and their abbreviations are:

| Attribute modifier | Abbreviation |
| ------------------ | ------------ |
| `after`            | `a`          |
| `before`           | `b`          |
| `empty`            | `e`          |
| `encrypted`        | *none*       |
| `exact`            | *none*       |
| `executable`       | `x`          |
| `external`         | *none*       |
| `once`             | `o`          |
| `onchange`         | *none*       |
| `private`          | `p`          |
| `readonly`         | `r`          |
| `remove`           | *none*       |
| `template`         | `t`          |

The type of a target can be changed using a type modifier:

| Type modifier |
| ------------- |
| `create`      |
| `modify`      |
| `script`      |
| `symlink`     |

The negative form of type modifiers, e.g. `nocreate`, changes the target to be
a regular file if it is of that type, otherwise the type is left unchanged.

Multiple modifications may be specified by separating them with a comma (`,`).
If you use the `-`*modifier* form then you must put *modifier* after a `--` to
prevent chezmoi from interpreting `-`*modifier* as an option.

## Common flags

### `-r`, `--recursive`

--8<-- "common-flags/recursive.md:default-false"

## Examples

```sh
chezmoi chattr template ~/.bashrc
chezmoi chattr noempty ~/.profile
chezmoi chattr private,template ~/.netrc
chezmoi chattr -- -x ~/.zshrc
chezmoi chattr +create,+private ~/.kube/config
```

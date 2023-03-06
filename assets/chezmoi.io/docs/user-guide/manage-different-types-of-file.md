# Manage different types of file

## Have chezmoi create a directory, but ignore its contents

If you want chezmoi to create a directory, but ignore its contents, say
`~/src`, first run:

```console
$ mkdir -p $(chezmoi source-path)/src
```

This creates the directory in the source state, which means that chezmoi will
create it (if it does not already exist) when you run `chezmoi apply`.

However, as this is an empty directory it will be ignored by git. So, create a
file in the directory in the source state that will be seen by git (so git does
not ignore the directory) but ignored by chezmoi (so chezmoi does not include it
in the target state):

```console
$ touch $(chezmoi source-path)/src/.keep
```

chezmoi automatically creates `.keep` files when you add an empty directory
with `chezmoi add`.

## Ensure that a target is removed

Create a file called `.chezmoiremove` in the source directory containing a list
of patterns of files to remove. chezmoi will remove anything in the target
directory that matches the pattern. As this command is potentially dangerous,
you should run chezmoi in verbose, dry-run mode beforehand to see what would be
removed:

```console
$ chezmoi apply --dry-run --verbose
```

`.chezmoiremove` is interpreted as a template, so you can remove different
files on different machines. Negative matches (patterns prefixed with a `!`) or
targets listed in `.chezmoiignore` will never be removed.

## Manage part, but not all, of a file

chezmoi, by default, manages whole files, but there are two ways to manage just
parts of a file.

Firstly, a `modify_` script receives the current contents of the file on the
standard input and chezmoi reads the target contents of the file from the
script's standard output. This can be used to change parts of a file, for
example using `sed`.

!!! hint

    If you need random access to the file to modify it, then you can write
    standard input to a temporary file, modify the temporary file, and then
    write the temporary file to the standard output, for example:

    ```sh
    #!/bin/sh
    tempfile="$(mktemp)"
    trap 'rm -rf "${tempfile}"' EXIT
    cat > "${tempfile}"
    # modify ${tempfile}
    cat "${tempfile}"
    ```

!!! note

    If the file does not exist then the standard input to the `modify_` script
    will be empty and it is the script's responsibility to write a complete
    file to the standard output.

`modify_` scripts that contain the string `chezmoi:modify-template` are
executed as templates with the current contents of the file passed as
`.chezmoi.stdin` and the result of the template execution used as the new
contents of the file.

!!! example

    To replace the string `old` with `new` in a file while leaving the rest of
    the file unchanged, use the modify script:

    ```
    {{- /* chezmoi:modify-template */ -}}
    {{- .chezmoi.stdin | replaceAllRegex "old" "new" }}
    ```

    To set individual values in JSON, JSONC, TOML, and YAML files you can use
    the `setValueAtPath` template function, for example:

    ```
    {{- /* chezmoi:modify-template */ -}}
    {{ fromJson .chezmoi.stdin | setValueAtPath "key.nestedKey" "value" | toPrettyJson }}
    ```

!!! warning

    Modify templates must not have a `.tmpl` extension.

Secondly, if only a small part of the file changes then consider using a
template to re-generate the full contents of the file from the current state.
For example, Kubernetes configurations include a current context that can be
substituted with:

``` title="~/.local/share/chezmoi/dot_kube/config.tmpl"
current-context: {{ output "kubectl" "config" "current-context" | trim }}
```

!!! hint

    For managing ini files with a mix of settings and state (such as recently
    used files or window positions), there is a third party tool called
    `chezmoi_modify_manager` that builds upon `modify_` scripts. See
    [related software](../links/related-software.md#githubcomvorpalbladechezmoi_modify_manager)
    for more information.


## Manage a file's permissions, but not its contents

chezmoi's `create_` attributes allows you to tell chezmoi to create a file if
it does not already exist. chezmoi, however, will apply any permission changes
from the `executable_`, `private_`, and `readonly_` attributes. This can be
used to control a file's permissions without altering its contents.

For example, if you want to ensure that `~/.kube/config` always has permissions
600 then if you create an empty file called `dot_kube/private_config` in
your source state, chezmoi will ensure `~/.kube/config`'s permissions are 0600
when you run `chezmoi apply` without changing its contents.

This approach does have the downside that chezmoi will create the file if it
does not already exist. If you only want `chezmoi apply` to set a file's
permissions if it already exists and not create the file otherwise, you can use
a `run_` script. For example, create a file in your source state called
`run_set_kube_config_permissions.sh` containing:

```bash
#!/bin/sh

FILE="$HOME/.kube/config"
if [ -f "$FILE" ]; then
    if [ "$(stat -c %a "$FILE")" != "600" ] ; then
        chmod 600 "$FILE"
    fi
fi
```

## Handle configuration files which are externally modified

Some programs modify their configuration files. When you next run `chezmoi
apply`, any modifications made by the program will be lost.

You can track changes to these files by replacing with a symlink back to a file
in your source directory, which is under version control. Here is a worked
example for VSCode's `settings.json` on Linux:

Copy the configuration file to your source directory:

```console
$ cp ~/.config/Code/User/settings.json $(chezmoi source-path)
```

Tell chezmoi to ignore this file:

```console
$ echo settings.json >> $(chezmoi source-path)/.chezmoiignore
```

Tell chezmoi that `~/.config/Code/User/settings.json` should be a symlink to
the file in your source directory:

```console
$ mkdir -p $(chezmoi source-path)/private_dot_config/private_Code/User
$ echo -n "{{ .chezmoi.sourceDir }}/settings.json" > $(chezmoi source-path)/private_dot_config/private_Code/User/symlink_settings.json.tmpl
```

The prefix `private_` is used because the `~/.config` and `~/.config/Code`
directories are private by default.

Apply the changes:

```console
$ chezmoi apply -v
```

Now, when the program modifies its configuration file it will modify the file
in the source state instead.

## Populate `~/.ssh/authorized_keys` with your public SSH keys from GitHub

chezmoi can retrieve your public SSH keys from GitHub, which can be useful for
populating your `~/.ssh/authorized_keys`. Put the following in your
`~/.local/share/chezmoi/dot_ssh/authorized_keys.tmpl`:

```
{{ range gitHubKeys "$GITHUB_USERNAME" -}}
{{   .Key }}
{{ end -}}
```

# Target types

chezmoi will create, update, and delete files, directories, and symbolic links
in the destination directory, and run scripts. chezmoi deterministically
performs actions in ASCII order of their target name.

!!! example

    Given a file `dot_a`, a script `run_z`, and a directory `exact_dot_c`, chezmoi
    will first create `.a`, create `.c`, and then execute `run_z`.

## Files

Files are represented by regular files in the source state. The `encrypted_`
attribute determines whether the file in the source state is encrypted. The
`executable_` attribute will set the executable bits in the target state,
and the `private_` attribute will clear all group and world permissions. The
`readonly_` attribute will clear all write permission bits in the target state.
Files with the `.tmpl` suffix will be interpreted as templates. If the target
contents are empty then the file will be removed, unless it has an `empty_`
prefix.

### Create file

Files with the `create_` prefix will be created in the target state with the
contents of the file in the source state if they do not already exist. If the
file in the destination state already exists then its contents will be left
unchanged.

### Modify file

Files with the `modify_` prefix are treated as scripts that modify an existing
file.

If the file contains a line with the text `chezmoi:modify-template` then that
line is removed and the rest of the script is executed template with the
existing file's contents passed as a string in `.chezmoi.stdin`. The result of
executing the template are the new contents of the file.

Otherwise, the contents of the existing file (which maybe empty if the existing
file does not exist or is empty) are passed to the script's standard input, and
the new contents are read from the script's standard output.

### Remove entry

Files with the `remove_` prefix will cause the corresponding entry (file,
directory, or symlink) to be removed in the target state.

## Directories

Directories are represented by regular directories in the source state. The
`exact_` attribute causes chezmoi to remove any entries in the target state that
are not explicitly specified in the source state, and the `private_` attribute
causes chezmoi to clear all group and world permissions. The `readonly_`
attribute will clear all write permission bits.

## Symbolic links

Symbolic links are represented by regular files in the source state with the
prefix `symlink_`. The contents of the file will have a trailing newline
stripped, and the result be interpreted as the target of the symbolic link.
Symbolic links with the `.tmpl` suffix in the source state are interpreted as
templates. If the target of the symbolic link is empty or consists only of
whitespace, then the target is removed.

## Scripts

Scripts are represented as regular files in the source state with prefix `run_`.
The file's contents (after being interpreted as a template if it has a `.tmpl`
suffix) are executed.

Scripts are executed on every `chezmoi apply`, unless they have the `once_` or
`onchange_` attribute. `run_once_` scripts are only executed if a script with
the same contents has not been run before, i.e. if the script is new or if its
contents have changed. `run_onchange_` scripts are executed whenever their
contents change, even if a script with the same contents has run before.

Scripts with the `before_` attribute are executed before any files, directories,
or symlinks are updated. Scripts with the `after_` attribute are executed after
all files, directories, and symlinks have been updated. Scripts without an
`before_` or `after_` attribute are executed in ASCII order of their target
names with respect to files, directories, and symlinks.

Scripts will normally run with their working directory set to their equivalent
location in the destination directory. If the equivalent location in the
destination directory either does not exist or is not a directory, then chezmoi
will walk up the script's directory hierarchy and run the script in the first
directory that exists and is a directory.

!!! example

    A script in `~/.local/share/chezmoi/dir/run_script` will be run with a working
    directory of `~/dir`.

The `scriptEnv` configuration variable specifies extra environment variables
when running the script.

### Scripts on Windows

<!-- FIXME: some of the following needs to be moved to the how-to -->

The execution of scripts on Windows depends on the script's file extension.
Windows will natively execute scripts with a `.bat`, `.cmd`, `.com`, and `.exe`
extensions. Other extensions require an interpreter, which must be in your
`%PATH%`.

The default script interpreters are:

| Extension | Command      | Arguments |
| --------- | ------------ | --------- |
| `.pl`     | `perl`       | *none*    |
| `.py`     | `python3`    | *none*    |
| `.ps1`    | `powershell` | `-NoLogo` |
| `.rb`     | `ruby`       | *none*    |

Script interpreters can be added or overridden by adding the corresponding
extension (without the leading dot) as a key under the `interpreters`
section of the configuration file.

!!! example

    To change the Python interpreter to `C:\Python39\python3.exe` and add a
    Tcl/Tk interpreter, include the following in your config file:

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [interpreters]
    py = { command = 'C:\Python39\python3.exe' }
    tcl = { command = "tclsh" }
    ```

    If using TOML, it's also possible to write it like this:
    
    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [interpreters.py]
        command = 'C:\Python39\python3.exe'
    [interpreters.tcl]
        command = "tclsh"
    ```

    Or if using YAML:

    ```yaml title="~/.config/chezmoi/chezmoi.yaml"
    interpreters:
      py:
        command: "C:\Python39\python3.exe"
      tcl:
        command: "tclsh"
    ```

!!! note

    If you intend to use PowerShell Core (`pwsh.exe`) as the `.ps1`
    interpreter, include the following in your config file:

    ```toml title="~/.confg/chezmoi/chezmoi.toml"
    [interpreters.ps1]
        command = "pwsh"
        args = ["-NoLogo"]
    ```

If the script in the source state is a template (with a `.tmpl` extension), then
chezmoi will strip the `.tmpl` extension and use the next remaining extension to
determine the interpreter to use.

## `symlink` mode

By default, chezmoi will create regular files and directories. Setting `mode =
"symlink"` will make chezmoi behave more like a dotfile manager that uses
symlinks by default, i.e. `chezmoi apply` will make dotfiles symlinks to files
in the source directory if the target is a regular file and is not
encrypted, executable, private, or a template.

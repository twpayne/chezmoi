# Interpreters

<!-- FIXME: some of the following needs to be moved to the how-to -->

The execution of scripts and hooks on Windows depends on the file extension.
Windows will natively execute scripts with a `.bat`, `.cmd`, `.com`, and `.exe`
extensions. Other extensions require an interpreter, which must be in your
`%PATH%`.

The default script interpreters are:

| Extension | Command      | Arguments       |
| --------- | ------------ | --------------- |
| `.nu`     | `nu`         | *none*          |
| `.pl`     | `perl`       | *none*          |
| `.py`     | `python3`    | *none*          |
| `.ps1`    | `pwsh`       | `-NoLogo -File` |
| `.rb`     | `ruby`       | *none*          |

Script interpreters can be added or overridden by adding the corresponding
extension (without the leading dot) as a key under the `interpreters`
section of the configuration file.

!!! note

    The leading `.` is dropped from *extension*, for example to specify the
    interpreter for `.pl` files you configure `interpreters.pl` (where `.`
    in this case just means "a child of" in the configuration file, however
    that is specified in your preferred format).

!!! example

    To change the Python interpreter to `C:\Python39\python3.exe` and add a
    Tcl/Tk interpreter, include the following in your config file:

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

    Note that the TOML version can also be written like this, which
    resembles the YAML version more and makes it clear that the key
    for each file extension should not have a leading `.`:

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [interpreters]
    py = { command = 'C:\Python39\python3.exe' }
    tcl = { command = "tclsh" }
    ```

!!! info "PowerShell Core Installation"

    PowerShell Core (`pwsh`) must be installed separately on most systems.
    If you don't have it installed:

    - **Windows**: Download from [PowerShell releases](https://github.com/PowerShell/PowerShell/releases)
      or install via `winget install Microsoft.PowerShell`
    - **Linux/macOS**: Follow instructions at [Installing PowerShell](https://learn.microsoft.com/powershell/scripting/install/installing-powershell)

!!! note

    chezmoi defaults to PowerShell Core (`pwsh`) for `.ps1` scripts as it
    provides better UTF-8 support and cross-platform compatibility. On Windows,
    if `pwsh` is not available, chezmoi will automatically fall back to Windows
    PowerShell (`powershell`).

    To explicitly use Windows PowerShell instead of the automatic selection,
    include the following in your config file:

    ```toml title="~/.config/chezmoi/chezmoi.toml"
    [interpreters.ps1]
        command = "powershell"
        args = ["-NoLogo"]
    ```

If the script in the source state is a template (with a `.tmpl` extension), then
chezmoi will strip the `.tmpl` extension and use the next remaining extension to
determine the interpreter to use.

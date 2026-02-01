# Windows

## Detect Windows Subsystem for Linux (WSL)

WSL can be detected by looking for the string `Microsoft` or `microsoft` in
`/proc/sys/kernel/osrelease`, which is available in the template variable
`.chezmoi.kernel.osrelease`, for example:

```text
{{ if eq .chezmoi.os "linux" }}
{{   if (.chezmoi.kernel.osrelease | lower | contains "microsoft") }}
# WSL-specific code
{{   end }}
{{ end }}
```

## Run a PowerShell script as admin on Windows

Put the following at the top of your script:

```powershell
# Self-elevate the script if required
if (-Not ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] 'Administrator')) {
  if ([int](Get-CimInstance -Class Win32_OperatingSystem | Select-Object -ExpandProperty BuildNumber) -ge 6000) {
    $CommandLine = "-NoExit -File `"" + $MyInvocation.MyCommand.Path + "`" " + $MyInvocation.UnboundArguments
    Start-Process -Wait -FilePath pwsh.exe -Verb Runas -ArgumentList $CommandLine
    Exit
  }
}
```

!!! tip

    The example above uses `pwsh.exe` (PowerShell Core). If you don't have
    PowerShell Core installed or prefer Windows PowerShell, change `pwsh.exe`
    to `PowerShell.exe` in the `Start-Process` command.

If you use [gsudo][gsudo], it has tips on writing [self-elevating scripts][ses].

## Run elevated scripts

However you decide to run a script in an elevated prompt, as soon as the
non-elevated script returns, chezmoi will move to the next step in its
processing (running more scripts, creating files, etc.). Ensure that the
elevated script completes *before the non-elevated script exits*, or subsequent
steps may not run as expected. In the example above, this is accomplished by
passing `-Wait` to PowerShell's `Start-Process` cmdlet.

Note that by including `-NoExit` in `$CommandLine`, the new (elevated)
PowerShell process/window will not exit automatically on completion. This means
you'll need to close the new window by hand for chezmoi to continue its steps.
If this manual intervention is desired, it would be convenient to print a
message as the script's last command to indicate completion for you to safely
close the elevated window. If you want no manual intervention, you can remove
`-NoExit` from `$CommandLine`, but then you likely won’t see the output of the
elevated script, which will make it more difficult to determine if something
went wrong during its execution.

## Create symlinks

For chezmoi to be able to create symlinks, the current user must have the
appropriate permissions. This can be achieved by turning on [developer
mode][devmode] or setting the permission `SeCreateSymbolicLinkPrivilege`
[directly][symperm].

[gsudo]: https://gerardog.github.io/gsudo/docs/intro
[ses]: https://gerardog.github.io/gsudo/docs/tips/script-self-elevation#self-elevate-script
[devmode]: https://learn.microsoft.com/en-us/windows/advanced-settings/developer-mode
[symperm]: https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-10/security/threat-protection/security-policy-settings/create-symbolic-links

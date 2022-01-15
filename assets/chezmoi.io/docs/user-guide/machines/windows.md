# Windows

## Detect Windows Subsystem for Linux (WSL)

WSL can be detected by looking for the string `Microsoft` or `microsoft` in
`/proc/sys/kernel/osrelease`, which is available in the template variable
`.chezmoi.kernel.osrelease`, for example:

```
{{ if (eq .chezmoi.os "linux") }}
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
    Start-Process -FilePath PowerShell.exe -Verb Runas -ArgumentList $CommandLine
    Exit
  }
}
```

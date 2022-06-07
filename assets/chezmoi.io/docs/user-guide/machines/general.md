# General

## Determine whether the current machine is a laptop or desktop

The following template sets the `$chassisType` variable to `"desktop"` or
`"laptop"` on macOS, Linux, and Windows.

```
{{- $chassisType := "desktop" }}
{{- if eq .chezmoi.os "darwin" }}
{{-   if contains "MacBook" (output "sysctl" "-n" "hw.model") }}
{{-     $chassisType = "laptop" }}
{{-   else }}
{{-     $chassisType = "desktop" }}
{{-   end }}
{{- else if eq .chezmoi.os "linux" }}
{{-   $chassisType = (output "hostnamectl" "--json=short" | mustFromJson).Chassis }}
{{- else if eq .chezmoi.os "windows" }}
{{-   $chassisType = (output "powershell.exe" "-NoProfile" "-NonInteractive" "-Command" "if ((Get-CimInstance -Class Win32_Battery | Measure-Object).Count -gt 0) { Write-Output 'laptop' } else { Write-Output 'desktop' }") | trim }}
{{- end }}
```

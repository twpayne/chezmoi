# General

## Determine whether the current machine is a laptop or desktop

The following template sets the `$chassisType` variable to `"desktop"` or
`"laptop"` on macOS, Linux, and Windows.

```text
{{- $chassisType := "desktop" }}
{{- if eq .chezmoi.os "darwin" }}
{{-   if contains "MacBook" (output "system_profiler" "SPHardwareDataType") }}
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

## Determine how many CPU cores and threads the current machine has

The following template sets the `$cpuCores` and `$cpuThreads` variables to the
number of CPU cores and threads on the current machine respectively on
macOS, Linux and Windows.

```text
{{- $cpuCores := 1 }}
{{- $cpuThreads := 1 }}
{{- if eq .chezmoi.os "darwin" }}
{{-   $cpuCores = (output "sysctl" "-n" "hw.physicalcpu_max") | trim | atoi }}
{{-   $cpuThreads = (output "sysctl" "-n" "hw.logicalcpu_max") | trim | atoi }}
{{- else if eq .chezmoi.os "linux" }}
{{-   $cpuCores = (output "sh" "-c" "lscpu --online --parse | grep --invert-match '^#' | sort --field-separator=',' --key='2,4' --unique | wc --lines") | trim | atoi }}
{{-   $cpuThreads = (output "sh" "-c" "lscpu --online --parse | grep --invert-match '^#' | wc --lines") | trim | atoi }}
{{- else if eq .chezmoi.os "windows" }}
{{-   $cpuCores = (output "powershell.exe" "-NoProfile" "-NonInteractive" "-Command" "(Get-CimInstance -ClassName 'Win32_Processor').NumberOfCores") | trim | atoi }}
{{-   $cpuThreads = (output "powershell.exe" "-NoProfile" "-NonInteractive" "-Command" "(Get-CimInstance -ClassName 'Win32_Processor').NumberOfLogicalProcessors") | trim | atoi }}
{{- end }}
```

!!! example

    ```text title="~/.local/share/chezmoi/.chezmoi.toml.tmpl"
    [data.cpu]
    cores = {{ $cpuCores }}
    threads = {{ $cpuThreads }}
    ```

    ```text title="~/.local/share/chezmoi/is_hyperthreaded.txt.tmpl"
    {{- if gt .cpu.threads .cpu.cores -}}
    Hyperthreaded!
    {{- else -}}
    Not hyperthreaded!
    {{- end -}}
    ```

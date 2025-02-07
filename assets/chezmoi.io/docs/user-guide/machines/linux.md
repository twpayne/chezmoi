# Linux

## Combine operating system and Linux distribution conditionals

There can be as much variation between Linux distributions as there is between
operating systems. Due to `text/template`'s eager evaluation of conditionals,
this means you often have to write templates with nested conditionals:

```text
{{ if eq .chezmoi.os "darwin" }}
# macOS-specific code
{{ else if eq .chezmoi.os "linux" }}
{{   if eq .chezmoi.osRelease.id "debian" }}
# Debian-specific code
{{   else if eq .chezmoi.osRelease.id "fedora" }}
# Fedora-specific code
{{   end }}
{{ end }}
```

This can be simplified by combining the operating system and distribution into a
single custom template variable. Put the following in your configuration file
template:

```text
{{- $osid := .chezmoi.os -}}
{{- if hasKey .chezmoi.osRelease "id" -}}
{{-   $osid = printf "%s-%s" .chezmoi.os .chezmoi.osRelease.id -}}
{{- end -}}

[data]
    osid = {{ $osid | quote }}
```

This defines the `.osid` template variable to be `{{ .chezmoi.os }}` on machines
without an [`os-release` file][os-release], or to be `{{ .chezmoi.os }}-{{
.chezmoi.osRelease.id }}` on machines with an `os-release` file.

You can then simplify your conditionals to be:

```text
{{ if eq .osid "darwin" }}
# macOS-specific code
{{ else if eq .osid "linux-debian" }}
# Debian-specific code
{{ else if eq .osid "linux-fedora" }}
# Fedora-specific code
{{ end }}
```

[os-release]: https://www.freedesktop.org/software/systemd/man/os-release.html

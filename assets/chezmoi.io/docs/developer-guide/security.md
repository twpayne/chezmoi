# Security

## Supported versions

Only the most recent version of chezmoi is supported with security updates.

## Virus scanner false positives

Virus scanning software, especially on Windows machines, occasionally report
viruses or trojans in the chezmoi binary. This is almost certainly a false
positive.

For more information see [Why does my virus-scanning software think my Go
distribution or compiled binary is infected?][false] in
the Go FAQ.

## Reporting a vulnerability

Please report vulnerabilities by [opening a GitHub issue][issue] or sending an
email to [`twpayne+chezmoi-security@gmail.com`][email].

## Security thanks

Security problems are reported privately and do not show up in the list of
contributors. chezmoi thanks:

### Organizations

* [Secur0](https://secur0.com/) who performed an extensive security analysis of
  chezmoi.

### Individuals

* eltiburon7
* gueco
* julichaan
* Krypt3d (multiple vulnerabilities)
* pulpo
* reddev (multiple vulnerabilities)
* TheRedP4nther (multiple vulnerabilities)

[false]: https://go.dev/doc/faq#virus
[issue]: https://github.com/twpayne/chezmoi/issues/new/choose
[email]: mailto:twpayne%2Bchezmoi-security@gmail.com

# Comparison table

[chezmoi]: https://chezmoi.io/
[dotbot]: https://github.com/anishathalye/dotbot
[rcm]: https://github.com/thoughtbot/rcm
[homesick]: https://github.com/technicalpickles/homesick
[vcsh]: https://github.com/RichiH/vcsh
[yadm]: https://yadm.io/
[bare git]: https://www.atlassian.com/git/tutorials/dotfiles "bare git"

|                                        | [chezmoi]     | [dotbot]          | [rcm]             | [homesick]        | [vcsh]                   | [yadm]                       | [bare git] |
| -------------------------------------- | ------------- | ----------------- | ----------------- | ----------------- | ------------------------ | ---------------------------- | ---------- |
| Distribution                           | Single binary | Python package    | Multiple files    | Ruby gem          | Single script or package | Single script                | -          |
| Install method                         | Many          | git submodule     | Many              | Ruby gem          | Many                     | Many                         | Manual     |
| Non-root install on bare system        | ✅            | ⁉️                 | ⁉️                 | ⁉️                 | ✅                       | ✅                           | ✅         |
| Windows support                        | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅                           | ✅         |
| Bootstrap requirements                 | None          | Python, git       | Perl, git         | Ruby, git         | sh, git                  | git                          | git        |
| Source repos                           | Single        | Single            | Multiple          | Single            | Multiple                 | Single                       | Single     |
| dotfiles are...                        | Files         | Symlinks          | Files             | Symlinks          | Files                    | Files                        | Files      |
| Config file                            | Optional      | Required          | Optional          | None              | None                     | Optional                     | Optional   |
| Private files                          | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅                           | ❌         |
| Show differences without applying      | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅                           | ✅         |
| Whole file encryption                  | ✅            | ❌                | ❌                | ❌                | ❌                       | ✅                           | ❌         |
| Password manager integration           | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌                           | ❌         |
| Machine-to-machine file differences    | Templates     | Alternative files | Alternative files | Alternative files | Branches                 | Alternative files, templates | ⁉️          |
| Custom variables in templates          | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌                           | ❌         |
| Executable files                       | ✅            | ✅                | ✅                | ✅                | ✅                       | ✅                           | ✅         |
| File creation with initial contents    | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌                           | ❌         |
| Externals                              | ✅            | ❌                | ❌                | ❌                | ❌                       | ❌                           | ❌         |
| Manage partial files                   | ✅            | ❌                | ❌                | ❌                | ⁉️                        | ✅                           | ⁉️          |
| File removal                           | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅                           | ❌         |
| Directory creation                     | ✅            | ✅                | ✅                | ❌                | ✅                       | ✅                           | ✅         |
| Run scripts                            | ✅            | ✅                | ✅                | ❌                | ✅                       | ✅                           | ❌         |
| Run once scripts                       | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅                           | ❌         |
| Machine-to-machine symlink differences | ✅            | ❌                | ❌                | ❌                | ⁉️                        | ✅                           | ⁉️          |
| Shell completion                       | ✅            | ❌                | ❌                | ❌                | ✅                       | ✅                           | ✅         |
| Archive import                         | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌                           | ✅         |
| Archive export                         | ✅            | ❌                | ❌                | ❌                | ✅                       | ❌                           | ✅         |
| Implementation language                | Go            | Python            | Perl              | Ruby              | POSIX Shell              | Bash                         | C          |

✅ Supported, ⁉️  Possible with significant manual effort, ❌ Not supported

For more comparisons, visit [dotfiles.github.io](https://dotfiles.github.io/).

# Variables

chezmoi provides the following automatically-populated variables:

| Variable                      | Type     | Value                                                                                                                                                 |
|-------------------------------| -------- |-------------------------------------------------------------------------------------------------------------------------------------------------------|
| `.chezmoi.arch`               | string   | Architecture, e.g. `amd64`, `arm`, etc. as returned by [runtime.GOARCH](https://pkg.go.dev/runtime?tab=doc#pkg-constants)                             |
| `.chezmoi.args`               | []string | The arguments passed to the `chezmoi` command, starting with the program command                                                                      |
| `.chezmoi.cacheDir`           | string   | The cache directory                                                                                                                                   |
| `.chezmoi.config`             | object   | The configuration, as read from the config file                                                                                                       |
| `.chezmoi.configFile`         | string   | The path to the configuration file used by chezmoi                                                                                                    |
| `.chezmoi.executable`         | string   | The path to the `chezmoi` executable, if available                                                                                                    |
| `.chezmoi.fqdnHostname`       | string   | The fully-qualified domain name hostname of the machine chezmoi is running on                                                                         |
| `.chezmoi.gid`                | string   | The primary group ID                                                                                                                                  |
| `.chezmoi.group`              | string   | The group of the user running chezmoi                                                                                                                 |
| `.chezmoi.homeDir`            | string   | The home directory of the user running chezmoi                                                                                                        |
| `.chezmoi.hostname`           | string   | The hostname of the machine chezmoi is running on, up to the first `.`                                                                                |
| `.chezmoi.kernel`             | object   | Contains information from `/proc/sys/kernel`. Linux only, useful for detecting specific kernels (e.g. Microsoft's WSL kernel)                         |
| `.chezmoi.os`                 | string   | Operating system, e.g. `darwin`, `linux`, etc. as returned by [runtime.GOOS](https://pkg.go.dev/runtime?tab=doc#pkg-constants)                        |
| `.chezmoi.osRelease`          | object   | The information from `/etc/os-release`, Linux only, run `chezmoi data` to see its output                                                              |
| `.chezmoi.pathListSeparator`  | string   | The path list separator, typically `;` on Windows and `:` on other systems. Used to separate paths in environment variables. ie `/bin:/sbin:/usr/bin` |
| `.chezmoi.pathSeparator`      | string   | The path separator, typically `\` on windows and `/` on unix. Used to separate files and directories in a path. ie `c:\see\dos\run`                   |
| `.chezmoi.sourceDir`          | string   | The source directory                                                                                                                                  |
| `.chezmoi.sourceFile`         | string   | The path of the template relative to the source directory                                                                                             |
| `.chezmoi.targetFile`         | string   | The absolute path of the target file for the template                                                                                                 |
| `.chezmoi.uid`                | string   | The user ID                                                                                                                                           |
| `.chezmoi.username`           | string   | The username of the user running chezmoi                                                                                                              |
| `.chezmoi.version.builtBy`    | string   | The program that built the `chezmoi` executable, if set                                                                                               |
| `.chezmoi.version.commit`     | string   | The git commit at which the `chezmoi` executable was built, if set                                                                                    |
| `.chezmoi.version.date`       | string   | The timestamp at which the `chezmoi` executable was built, if set                                                                                     |
| `.chezmoi.version.version`    | string   | The version of chezmoi                                                                                                                                |
| `.chezmoi.windowsVersion`     | object   | Windows version information, if running on Windows                                                                                                    |
| `.chezmoi.workingTree`        | string   | The working tree of the source directory                                                                                                              |

`.chezmoi.windowsVersion` contains the following keys populated from the
registry key `Computer\HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows
NT\CurrentVersion`.

| Key                         | Type    |
| --------------------------- | ------- |
| `currentBuild`              | string  |
| `currentMajorVersionNumber` | integer |
| `currentMinorVersionNumber` | integer |
| `currentVersion`            | string  |
| `displayVersion`            | string  |
| `editionID`                 | string  |
| `productName`               | string  |

Additional variables can be defined in the config file in the `data` section.
Variable names must consist of a letter and be followed by zero or more letters
and/or digits.

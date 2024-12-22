# `.chezmoiscripts`

If a directory called `.chezmoiscripts` exists in the root of the source
directory then any scripts in it are executed as normal scripts without
creating a corresponding directory in the target state.

## OS related scripts

If you have scripts that should only be executed on specific operating systems, you can avoid wrapping all instructions in templates by placing them in the following subdirectories within `.chezmoiscripts`:

- `linux`: scripts in this directory will only be executed on Linux.
- `darwin`: scripts in this directory will only be executed on MacOS.
- `windows`: scripts in this directory will only be executed on Windows.

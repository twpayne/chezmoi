# Use chezmoi with Watchman

chezmoi can be used with [Watchman](https://facebook.github.io/watchman) to
automatically run `chezmoi apply` whenever your source state changes, but there
are some limitations because Watchman runs actions in the background without a
terminal.

Firstly, Watchman spawns a server which runs actions when filesystems change.
This server reads its environment variables when it is started, typically on the
first invocation of the `watchman` command. If you use a password manager that
uses environment variables to persist login sessions, then you must login to
your password manager before you run the first `watchman` command, and your
session might eventually time out.

Secondly, Watchman runs processes without a terminal, and so cannot run
interactive processes. For `chezmoi apply`, you can use the `--force` flag to
suppress prompts to overwrite files that have been modified since chezmoi last
wrote them. However, if any other part of `chezmoi apply` is interactive, for
example if your password manager prompts for a password, then it will not work
with Watchman.

1. Tell watchman to watch your source directory:

    ```sh
    CHEZMOI_SOURCE_PATH="$(chezmoi source-path)"
    watchman watch "${CHEZMOI_SOURCE_PATH}"
    ```

2. Tell watchman to run `chezmoi apply --force` whenever your source directory
changes:

    ```sh
    watchman -j <<EOT
    ["trigger", "${CHEZMOI_SOURCE_PATH}", {
      "name": "chezmoi-apply",
      "command": ["chezmoi", "apply", "--force"]
    }]
    EOT
    ```

You can now make changes to your source directory and Watchman will run `chezmoi
apply --force` on each change.

To shutdown the Watchman server, run:

    ```sh
    watchman shutdown-server
    ```

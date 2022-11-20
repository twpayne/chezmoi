# `re-add`

Re-add all modified files in the target state. chezmoi will not overwrite
templates, and all entries that are not files are ignored.

!!! hint

    If you want to re-add a single file uncondtionally, use `chezmoi add --force` instead.

!!! example

    ```console
    $ chezmoi re-add
    ```

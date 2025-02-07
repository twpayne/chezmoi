# `ioreg`

On macOS, `ioreg` returns the structured output of the `ioreg -a -l` command,
which includes detailed information about the I/O Kit registry.

On non-macOS operating systems, `ioreg` returns `nil`.

The output from `ioreg` is cached so multiple calls to the `ioreg` function
will only execute the `ioreg -a -l` command once.

!!! example

    ```
    {{ if eq .chezmoi.os "darwin" }}
    {{   $serialNumber := index ioreg "IORegistryEntryChildren" 0 "IOPlatformSerialNumber" }}
    {{ end }}
    ```

!!! warning

    The `ioreg` function can be very slow and should not be used. It will be
    removed in a later version of chezmoi.

+/- 2.15.0

    Deprecated.

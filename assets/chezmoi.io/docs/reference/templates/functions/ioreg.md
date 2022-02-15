# `ioreg`

On macOS, `ioreg` returns the structured output of the `ioreg -a -l` command,
which includes detailed information about the I/O Kit registry.

On non-macOS operating systems, `ioreg` returns `nil`.

The output from `ioreg` is cached so multiple calls to the `ioreg` function
will only execute the `ioreg -a -l` command once.

!!! example

    ```
    {{ if (q .chezmoi.os "darwin" }}
    {{   $serialNumber := index ioreg "IORegistryEntryChildren" 0 "IOPlatformSerialNumber" }}
    {{ end }}
    ```

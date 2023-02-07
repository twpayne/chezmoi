# `toPrettyJson` [*indent*] *value*

`toPrettyJson` returns the JSON representation of *value*. The optional *indent*
specifies how much nested elements are indented relative to their parent.
*indent* defaults to two spaces.

!!! examples

    ```
    {{ dict "a" (dict "b" "c") | toPrettyJson "\t" }}
    ```

# `ensureLinePrefix` *prefix* [*prefix-to-add*] *text*

`ensureLinePrefix` ensures that each line of *text* starts with *prefix*. If any
line does not start with *prefix* then *prefix-to-add* is prepended to that
line.

Typically, `ensureLinePrefix` is used to ensure that lines are commented out,
similar to the [`comment` template function](comment.md). `ensureLinePrefix`
only modifies lines that are not already comments, whereas `comment` modifies
all lines, even if they are already comments.

!!! example

    ```
    {{ "### Heading\nBody\n" | ensureLinePrefix "#" }}
    {{ "### Heading\nBody\n" | ensureLinePrefix "#" "# " }}
    ```

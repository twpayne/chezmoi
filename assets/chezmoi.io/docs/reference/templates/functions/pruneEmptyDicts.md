# `pruneEmptyDicts` *dict*

`pruneEmptyDicts` modifies *dict* to remove nested empty dicts. Properties are
pruned from the bottom up, so any nested dicts that themselves only contain
empty dicts are pruned.

!!! example

    ```
    {{ dict "key" "value" "inner" (dict) | pruneEmptyDicts | toJson }}
    ```

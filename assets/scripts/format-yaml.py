#!/usr/bin/env python3

from __future__ import annotations

import sys
from pathlib import Path

from ruamel.yaml import YAML


def main() -> int:
    yaml = YAML()
    # ruamel.yaml.YAML will by default use the native line ending, which leads
    # to differences between Windows and UNIX. Force the output to use UNIX line
    # endings.
    yaml.line_break = '\n'
    # ruamel.yaml.YAML will by default break long lines and introduce trailing
    # whitespace errors. Disable this behavior by setting a long line width.
    yaml.width = 1024
    for filename in sys.argv[1:]:
        with Path(filename).open('r') as file:
            data = yaml.load(file)
        with Path(filename).open('w', newline='\n') as file:
            yaml.dump(data, file)
    return 0


if __name__ == '__main__':
    raise SystemExit(main())

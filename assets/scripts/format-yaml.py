#!/usr/bin/env python3

from __future__ import annotations

import sys
from pathlib import Path

from ruamel.yaml import YAML


def main() -> int:
    yaml = YAML()
    for filename in sys.argv[1:]:
        with Path(filename).open('r') as file:
            data = yaml.load(file)
        with Path(filename).open('w') as file:
            yaml.dump(data, file)
    return 0


if __name__ == '__main__':
    raise SystemExit(main())

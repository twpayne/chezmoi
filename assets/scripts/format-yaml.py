#!/usr/bin/env python3

import sys

from ruamel.yaml import YAML


def main():
    yaml = YAML()
    for filename in sys.argv[1:]:
        with open(filename) as file:
            data = yaml.load(file)
        with open(filename, 'w') as file:
            yaml.dump(data, file)


if __name__ == '__main__':
    main()

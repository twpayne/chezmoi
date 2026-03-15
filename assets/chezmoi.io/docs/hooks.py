from __future__ import annotations

import json
import re
import subprocess
import textwrap
import tomllib
from enum import Enum, auto
from io import StringIO
from pathlib import Path, PurePosixPath

from mkdocs import utils
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import Files
from ruamel.yaml import YAML

non_website_paths = [
    'docs.go',
    'hooks.py',
    'reference/commands/commands.go',
    'reference/commands/commands_test.go',
]

templates = [
    'index.md',
    'install.md',
    'links/articles.md',
    'links/podcasts.md',
    'links/videos.md',
    'reference/configuration-file/variables.md',
    'reference/release-history.md',
]


class MarkdownTransformState(Enum):
    NORMAL = auto()
    IN_BLOCK = auto()
    IN_FENCE = auto()


def build_code_fence(text: str, fmt: str, title: str | None = None) -> str:
    new_title = title
    if title:
        for suffix in ('.toml.tmpl', '.toml'):
            if title.endswith(suffix):
                new_title = title.removesuffix(suffix) + suffix.replace(
                    '.toml', f'.{fmt}'
                )
    return '\n'.join(
        (
            f"""```{fmt}{f' title="{new_title}"' if new_title else ''}""",
            text.strip(),
            '```',
        )
    )


def build_example_tabs(toml_text: str, title: str | None = None) -> str:
    data = tomllib.loads(toml_text)

    yaml_obj = YAML()
    yaml_obj.line_break = '\n'
    yaml_obj.width = 1024
    with StringIO() as yaml_stream:
        yaml_obj.dump(data, yaml_stream)
        yaml_text = yaml_stream.getvalue()

    json_text = json.dumps(data, indent=4)

    tabs = []
    for fmt, text in (
        ('toml', toml_text),
        ('yaml', yaml_text),
        ('json', json_text),
    ):
        # Indent each fence by 4 spaces in each tab, e.g.:
        # === "TOML"
        #
        #     ```toml
        fence = textwrap.indent(build_code_fence(text, fmt, title), ' ' * 4)
        tabs.append(f'=== "{fmt.upper()}"\n\n{fence}')

    return '\n\n'.join(tabs)


def transform_markdown(lines: list[str]) -> list[str]:  # noqa: PLR0912, PLR0915
    DIRECTIVE = 'example-formats'
    BLOCK_OPEN_RX = re.compile(rf'^(?P<indent>\s*)<!--\s*{DIRECTIVE}\s*-->\s*$')
    BLOCK_CLOSE_RX = re.compile(rf'^\s*<!--\s*/{DIRECTIVE}\s*-->\s*$')

    FENCE_OPEN_RX = re.compile(r'^\s*```toml(?:\s+title="(?P<title>[^"]+)")?\s*$')
    FENCE_CLOSE_RX = re.compile(r'^\s*```\s*$')

    output = []
    state = MarkdownTransformState.NORMAL

    open_directive_line = None
    indent = None
    title = None
    toml_lines = []
    seen_fence = False

    for i, line in enumerate(lines, start=1):
        match state:
            case MarkdownTransformState.NORMAL:
                if m := BLOCK_OPEN_RX.match(line):
                    state = MarkdownTransformState.IN_BLOCK
                    open_directive_line = i
                    indent = m.group('indent')
                    seen_fence = False
                elif BLOCK_CLOSE_RX.match(line):
                    raise ValueError(f'Unexpected closing directive at line {i}')
                else:
                    output.append(line)

            case MarkdownTransformState.IN_BLOCK:
                if BLOCK_OPEN_RX.match(line):
                    raise ValueError(f'Nested opening directive at line {i}')
                elif BLOCK_CLOSE_RX.match(line):
                    if not seen_fence:
                        raise ValueError(
                            f'No TOML fence in directive block beginning on line {open_directive_line}'  # noqa: E501
                        )
                    if not toml_lines:
                        raise ValueError(f'Empty TOML fence at line {i}')

                    toml_text = textwrap.dedent('\n'.join(toml_lines)).strip('\n')
                    examples = build_example_tabs(toml_text, title)
                    indented = textwrap.indent(examples, indent or '')

                    output.extend(indented.splitlines())

                    state = MarkdownTransformState.NORMAL
                    open_directive_line = None
                    indent = None
                    title = None
                    toml_lines = []
                    seen_fence = False
                elif not line.strip():
                    continue
                elif m := FENCE_OPEN_RX.match(line):
                    if seen_fence:
                        raise ValueError(f'Multiple TOML fences in block at line {i}')

                    state = MarkdownTransformState.IN_FENCE
                    title = m.group('title')
                    toml_lines = []
                    seen_fence = True
                else:
                    raise ValueError(
                        f'Unexpected content inside directive block at line {i}'
                    )
            case MarkdownTransformState.IN_FENCE:
                if BLOCK_OPEN_RX.match(line):
                    raise ValueError(f'Nested opening directive at line {i}')
                elif BLOCK_CLOSE_RX.match(line):
                    raise ValueError(
                        f'Unexpected closing directive before fence closed at line {i}'
                    )
                elif FENCE_CLOSE_RX.match(line):
                    state = MarkdownTransformState.IN_BLOCK
                else:
                    if line.lstrip().startswith('{{'):
                        raise ValueError(
                            f'Bare template syntax in TOML fence at line {i}'
                        )
                    toml_lines.append(line)

    if state is MarkdownTransformState.IN_BLOCK:
        raise ValueError('Unclosed directive block at EOF')
    if state is MarkdownTransformState.IN_FENCE:
        raise ValueError('Unclosed TOML fence at EOF')

    return output


def on_pre_build(config: MkDocsConfig, **kwargs) -> None:
    docs_dir = PurePosixPath(config.docs_dir)
    for src_path in templates:
        output_path = docs_dir.joinpath(src_path)
        template_path = output_path.parent / (output_path.name + '.tmpl')
        data_path = output_path.parent / (output_path.name + '.yaml')
        args = ['go', 'tool', 'execute-template']
        if Path(data_path).exists():
            args.extend(['-data', data_path])
        args.extend(['-output', output_path, template_path])
        subprocess.run(args, check=False)


def on_files(files: Files, **kwargs) -> Files:
    # remove non-website files
    for src_path in non_website_paths:
        files.remove(files.get_file_from_path(src_path))

    # remove templates and data
    for src_path in templates:
        files.remove(files.get_file_from_path(src_path + '.tmpl'))
        data_path = src_path + '.yaml'
        if data_path in files:
            files.remove(files.get_file_from_path(data_path))

    return files


def on_post_build(config: MkDocsConfig, **kwargs) -> None:
    config_dir = Path(config.config_file_path).parent
    site_dir = config.site_dir

    # copy GitHub pages config
    utils.copy_file(Path(config_dir, 'CNAME'), Path(site_dir, 'CNAME'))

    # copy installation scripts
    utils.copy_file(Path(config_dir, '../scripts/install.sh'), Path(site_dir, 'get'))
    utils.copy_file(
        Path(config_dir, '../scripts/install-local-bin.sh'),
        Path(site_dir, 'getlb'),
    )
    utils.copy_file(
        Path(config_dir, '../scripts/install.ps1'),
        Path(site_dir, 'get.ps1'),
    )

    # copy cosign.pub
    utils.copy_file(
        Path(config_dir, '../cosign/cosign.pub'),
        Path(site_dir, 'cosign.pub'),
    )


def on_page_markdown(markdown: str, **kwargs) -> str:
    lines = markdown.splitlines()

    return '\n'.join(transform_markdown(lines))

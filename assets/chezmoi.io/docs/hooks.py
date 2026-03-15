from __future__ import annotations

import json
import re
import subprocess
import textwrap
import tomllib
from io import StringIO
from pathlib import Path, PurePosixPath

from mkdocs import utils
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import Files
from ruamel.yaml import YAML
from ruamel.yaml.representer import RepresenterError as YAMLRepresenterError

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
    # matches TOML fences surrounded by a HTML comment-style directive,
    # captures the initial indentation, optional title filename and TOML text
    example_pattern = re.compile(
        r"""
        ^([ \t]*)
        <!--\s*example-formats\s*-->
        \s*
        ```toml(?:[ \t]+title="([^"]+(?:\.toml(?:\.tmpl)?)?)")?
        \s*
        (.*?)
        \s*
        ```
        \s*
        <!--\s*/example-formats\s*-->
        """,
        re.MULTILINE | re.DOTALL | re.VERBOSE,
    )

    def rename_with_format(filename_path: PurePosixPath, fmt: str) -> str:
        name = str(filename_path)
        for suffix in ('.toml.tmpl', '.toml'):
            if name.endswith(suffix):
                return name.removesuffix(suffix) + suffix.replace('.toml', f'.{fmt}')
        return name

    def build_code_fence(
        text: str, fmt: str, filename_path: PurePosixPath | None = None
    ) -> str:
        title = rename_with_format(filename_path, fmt) if filename_path else None

        return '\n'.join(
            (f"""```{fmt}{f' title="{title}"' if title else ''}""", text.strip(), '```')
        )

    def build_example_tabs(toml_text: str, filename: str | None = None) -> str:
        filename_path = PurePosixPath(filename) if filename else None

        data = tomllib.loads(textwrap.dedent(toml_text).strip())

        yaml_obj = YAML()
        yaml_obj.line_break = '\n'
        yaml_obj.width = 1024
        with StringIO() as yaml_stream:
            yaml_obj.dump(data, yaml_stream)
            yaml_text = yaml_stream.getvalue().strip()

        json_text = json.dumps(data, indent=4)

        blocks = []
        for fence, text in (
            ('toml', toml_text.strip()),
            ('yaml', yaml_text),
            ('json', json_text),
        ):
            block = textwrap.indent(
                build_code_fence(text, fence, filename_path), ' ' * 4
            )
            blocks.append(f'=== "{fence.upper()}"\n\n{block}')

        return '\n\n'.join(blocks)

    def replace(match: re.Match) -> str | None:
        indent = match.group(1)
        filename = match.group(2)
        toml_text = match.group(3)

        try:
            examples = build_example_tabs(toml_text, filename)
        except tomllib.TOMLDecodeError, YAMLRepresenterError, ValueError, TypeError:
            return match.group(0)

        return textwrap.indent(examples, indent)

    return example_pattern.sub(replace, markdown)

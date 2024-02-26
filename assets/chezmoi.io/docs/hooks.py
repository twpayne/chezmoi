from __future__ import annotations

import subprocess
from pathlib import Path, PurePosixPath

from mkdocs import utils
from mkdocs.config.defaults import MkDocsConfig
from mkdocs.structure.files import Files

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
    config_dir = Path(config.config_file_path).parent
    docs_dir = PurePosixPath(config.docs_dir)
    for src_path in templates:
        output_path = docs_dir.joinpath(src_path)
        template_path = output_path.parent / (output_path.name + '.tmpl')
        data_path = output_path.parent / (output_path.name + '.yaml')
        args = [
            'go',
            'run',
            Path(config_dir, '../../internal/cmds/execute-template/main.go'),
        ]
        if Path(data_path).exists():
            args.extend(['-data', data_path])
        args.extend(['-output', output_path, template_path])
        subprocess.run(args, check=False)


def on_files(files: Files, config: MkDocsConfig, **kwargs) -> Files:
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

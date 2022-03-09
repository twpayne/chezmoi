import os
import subprocess

from mkdocs import utils
from mkdocs.structure.files import File

non_website_paths = [
    "docs.go",
    "docs_test.go",
    "hooks.py",
    "reference/commands/commands.go",
    "reference/commands/commands_test.go",
]

templates = [
    "install.md",
    "links/articles-podcasts-and-videos.md",
]

def on_pre_build(config, **kwargs):
    docs_dir = config['docs_dir']
    for src_path in templates:
        output = docs_dir + "/" + src_path
        template = output + '.tmpl'
        data = output + '.yaml'
        subprocess.run(['go', 'run', '../../internal/cmds/execute-template', '-data', data, '-output', output, template])

def on_files(files, config, **kwargs):
    # remove non-website files
    for src_path in non_website_paths:
        files.remove(files.get_file_from_path(src_path))

    # remove templates and data
    for src_path in templates:
        files.remove(files.get_file_from_path(src_path + '.tmpl'))
        files.remove(files.get_file_from_path(src_path + '.yaml'))

    return files

def on_post_build(config, **kwargs):
    site_dir = config['site_dir']

    # copy GitHub pages config
    utils.copy_file('CNAME', os.path.join(site_dir, 'CNAME'))

    # copy installation scripts
    utils.copy_file('../scripts/install.sh', os.path.join(site_dir, 'get'))
    utils.copy_file('../scripts/install.ps1', os.path.join(site_dir, 'get.ps1'))

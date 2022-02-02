import os
import shutil

def on_post_build(config, **kwargs):
    site_dir = config['site_dir']

    # copy GitHub pages config
    shutil.copy('CNAME', os.path.join(site_dir, 'CNAME'))

    # copy installation scripts
    shutil.copy('../scripts/install.sh', os.path.join(site_dir, 'get'))
    shutil.copy('../scripts/install.ps1', os.path.join(site_dir, 'get.ps1'))

    # remove non-website files
    os.remove(os.path.join(site_dir, 'hooks.py'))
    os.remove(os.path.join(site_dir, 'reference/commands/commands.go'))
    os.remove(os.path.join(site_dir, 'reference/commands/commands_test.go'))

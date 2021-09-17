#!/bin/sh

for os in "$@"; do
    if [ -f "${os}.Vagrantfile" ]; then
        export VAGRANT_VAGRANTFILE=assets/vagrant/${os}.Vagrantfile
        if ( cd ../.. && vagrant up ); then
            vagrant ssh -c "sh test-chezmoi.sh"
            vagrant_ssh_exit_code=$?
            vagrant destroy -f || exit 1
            if [ $vagrant_ssh_exit_code -ne 0 ]; then
                exit $vagrant_ssh_exit_code
            fi
        else
            exit 1
        fi
    else
        echo "${os}.Vagrantfile not found"
        exit 1
    fi
done

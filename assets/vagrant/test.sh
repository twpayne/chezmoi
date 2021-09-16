#!/bin/sh

for os in "$@"; do
    if [ -f "${os}.Vagrantfile" ]; then
        export VAGRANT_VAGRANTFILE=assets/vagrant/${os}.Vagrantfile
        if ( cd ../.. && vagrant up ); then
            vagrant ssh -c "sh test-chezmoi.sh"
            vagrantSSHExitCode=$?
            vagrant destroy -f || exit 1
            if [ $vagrantSSHExitCode -ne 0 ]; then
                exit $vagrantSSHExitCode
            fi
        fi
    else
        echo "${os}.Vagrantfile not found"
    fi
done

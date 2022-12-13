#!/bin/bash

set -eufo pipefail

for os in "$@"; do
    echo "${os}"
    if [ ! -f "${os}.Vagrantfile" ]; then
        echo "${os}.Vagrantfile not found"
        exit 1
    fi
    export VAGRANT_VAGRANTFILE=assets/vagrant/${os}.Vagrantfile
    if ! ( cd ../.. && vagrant up ); then
        exit 1
    fi
    vagrant ssh -c "./test-chezmoi.sh"
    vagrant_ssh_exit_code=$?
    vagrant destroy -f || exit 1
    if [ $vagrant_ssh_exit_code -ne 0 ]; then
        exit $vagrant_ssh_exit_code
    fi
done

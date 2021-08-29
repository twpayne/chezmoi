#!/bin/sh

for os in "$@"; do
    if [ -f "${os}.Vagrantfile" ]; then
        export VAGRANT_VAGRANTFILE=assets/vagrant/${os}.Vagrantfile
        (
            cd ../.. &&
            vagrant up &&
            vagrant ssh -c "sh test-chezmoi.sh" &&
            vagrant destroy -f
        )
    else
        echo "${os}.Vagrantfile not found"
    fi
done

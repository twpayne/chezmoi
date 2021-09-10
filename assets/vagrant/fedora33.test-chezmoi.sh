#!/bin/sh

go get golang.org/dl/go1.17.1
"$HOME"/go/bin/go1.17.1 download
( cd /chezmoi && "$HOME"/go/bin/go1.17.1 test ./... )

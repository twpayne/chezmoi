#!/bin/sh

go get golang.org/dl/go1.17
"$HOME"/go/bin/go1.17 download
( cd /chezmoi && "$HOME"/go/bin/go1.17 test ./... )

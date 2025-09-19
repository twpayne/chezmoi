#!/bin/sh

set -euf

git config --global --add safe.directory /chezmoi

export GO="${GO:-go}"
export GOTOOLCHAIN=auto

if [ -d "/go-cache" ]; then
	export GOCACHE="/go-cache/cache"
	echo "Set GOCACHE to ${GOCACHE}"
	export GOMODCACHE="/go-cache/modcache"
	echo "Set GOMODCACHE to ${GOMODCACHE}"
fi

cd /chezmoi
${GO} tool chezmoi doctor || true
${GO} test ./...

sh assets/scripts/install.sh
bin/chezmoi --version

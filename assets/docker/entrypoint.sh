#!/bin/sh

set -euf

git config --global --add safe.directory /chezmoi

GO=${GO:-go}

if [ -d "/go-cache" ]; then
	export GOCACHE="/go-cache/cache"
	echo "Set GOCACHE to ${GOCACHE}"
	export GOMODCACHE="/go-cache/modcache"
	echo "Set GOMODCACHE to ${GOMODCACHE}"
fi

cd /chezmoi
${GO} run . doctor || true
${GO} test ./...

sh assets/scripts/install.sh
bin/chezmoi --version

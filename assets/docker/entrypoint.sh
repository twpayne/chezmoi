#!/bin/sh

set -euf

git config --global --add safe.directory /chezmoi

GO=${GO:-go}

if [ -d "/go-cache" ]; then
	echo "Set GOCACHE to /go-cache"
	export GOCACHE="/go-cache/cache"
	echo "Set GOMODCACHE to /go-cache/modcache"
	export GOMODCACHE="/go-cache/modcache"
fi

cd /chezmoi
${GO} run . doctor || true
${GO} test ./...

sh assets/scripts/install.sh
bin/chezmoi --version

#!/bin/bash

set -eufo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
REPO_ROOT=$(cd "${SCRIPT_DIR}/../.." &>/dev/null && pwd)
cd "${REPO_ROOT}"

if [ "$#" -eq 0 ]; then
	echo "Usage: $0 distribution1 [distribution2 ... distributionN]"
	exit 1
fi

for distribution in "$@"; do
	echo "${distribution}"
	dockerfile="assets/docker/${distribution}.Dockerfile"
	if [ ! -f "${dockerfile}" ]; then
		echo "${dockerfile} not found"
		exit 1
	fi
	image="$(docker build . -f "assets/docker/${distribution}.Dockerfile" -q)"
	docker_command=(
		docker run
		--env "CHEZMOI_GITHUB_ACCESS_TOKEN=${CHEZMOI_GITHUB_ACCESS_TOKEN-}"
		--env "CHEZMOI_GITHUB_TOKEN=${CHEZMOI_GITHUB_TOKEN-}"
		--env "GITHUB_ACCESS_TOKEN=${GITHUB_ACCESS_TOKEN-}"
		--env "GITHUB_TOKEN=${GITHUB_TOKEN-}"
		--rm
		--volume "${PWD}:/chezmoi"
	)
	if [ -n "${DOCKER_GOCACHE-}" ]; then
		mkdir -p "${DOCKER_GOCACHE}"
		docker_command+=(--volume "${DOCKER_GOCACHE}:/go-cache")
	fi
	docker_command+=("${image}")
	# Run docker
	"${docker_command[@]}"
done

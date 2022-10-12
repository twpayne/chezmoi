#!/bin/sh

output=""
while getopts "o:" arg; do
	case "${arg}" in
	o) output="${OPTARG}" ;;
	*) exit 1 ;;
	esac
done

commit="$(git rev-parse HEAD)"
if ! git diff-index --quiet HEAD; then
	commit="${commit}-dirty"
fi

if [ -z "${output}" ]; then
	echo "${commit}"
else
	echo "${commit}" > "${output}"
fi

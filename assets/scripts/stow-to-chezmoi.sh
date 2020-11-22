#!/usr/bin/env bash

set -e

BASEDIR="${1:-$HOME}"
STOWDIR="${2:-dotfiles}"

BASEDIR="$(
	unset CDPATH
	cd "${BASEDIR}" >/dev/null 2>&1
	pwd
)"

# if we have greadlink, use that
READLINK="$(which greadlink 2>/dev/null || which readlink)"

removelink() {
	[ -L "$1" ] && (
		LINK_DEST="$($READLINK -f "$1")"
		rm "$1"
		echo -ne "${LINK_DEST} ==> $1\t"
		if cp -R "${LINK_DEST}" "$1"; then
			echo "Done"
		else
			echo "FAILED"
			return 1
		fi
	)
}

work_file="$(mktemp)"
act_file="$(mktemp)"

trap 'rm -f ${work_file} ${act_file}' EXIT

find "${BASEDIR}" -not -path "${BASEDIR}/${STOWDIR}*" -type l >"${work_file}" || echo "Find skipped some files"

while read -r f; do
	target="$($READLINK -f "${f}" || echo '')"
	if [[ "${target}" == "${BASEDIR}/${STOWDIR}/"* ]]; then
		echo "Add $f"
		echo "${f}" >>"${act_file}"
	fi
done <"${work_file}"

read -p "Migrate the above to chezmoi? y/N" -r migrate
case "${migrate}" in
[Yy]*)
	echo "Migrating..."
	;;
*) exit 1 ;;
esac

mkdir -p "${BASEDIR}/.local/share"

while read -r f; do
	if removelink "${f}"; then
		chezmoi --source "${BASEDIR}/.local/share/chezmoi" --destination "${BASEDIR}" add "${f}"
	else
		echo "Unable to move: $f"
	fi
done <"${act_file}"

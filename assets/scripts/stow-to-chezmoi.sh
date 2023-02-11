#!/bin/sh

set -e

BASEDIR="${1:-${HOME}}"
STOWDIR="${2:-dotfiles}"

BASEDIR="$(
	unset -v CDPATH
	cd -- "${BASEDIR}" >/dev/null 2>&1
	pwd && printf .
)"
BASEDIR="${BASEDIR%??}"

# if we have greadlink, use that
READLINK="$(command -v greadlink 2>/dev/null || command -v readlink)"

removelink() {
	[ -h "${1}" ] && (
		LINK_DEST="$("${READLINK}" -f -- "${1}" && printf .)"
		LINK_DEST="${LINK_DEST%??}"
		rm -- "${1}"
		printf '%s ==> %s\t' "${LINK_DEST}" "${1}" >&2
		if cp -r -- "${LINK_DEST}" "${1}"; then
			printf 'Done\n' >&2
		else
			printf 'FAILED\n' >&2
			exit 1
		fi
	)
}

work_file="$(mktemp)"
act_file="$(mktemp)"

# attempt to clean up temporary files on exit
trap 'rm -f -- "${work_file}" "${act_file}"' EXIT
trap 'exit' INT TERM

find "${BASEDIR}" \! -path '*
*' \! -path "${BASEDIR}/${STOWDIR}*" -type l >"${work_file}" ||
	printf "Find skipped some files\n" >&2

while read -r f; do
	target="$("${READLINK}" -f -- "${f}" || :)"
	case "${target}" in
	"${BASEDIR}/${STOWDIR}/"*)
		printf 'Add %s\n' "${f}" >&2
		printf '%s\n' "${f}" >>"${act_file}"
		;;
	esac
done <"${work_file}"

printf 'Migrate the above to chezmoi? y/N ' >&2
read -r migrate
case "${migrate}" in
[Yy]*) printf 'Migrating...\n' >&2 ;;
*) exit 1 ;;
esac

mkdir -p -- "${BASEDIR}/.local/share"

while read -r f; do
	if removelink "${f}"; then
		chezmoi --source "${BASEDIR}/.local/share/chezmoi" --destination "${BASEDIR}" add -- "${f}"
	else
		printf 'Unable to move: %s\n' "${f}" >&2
	fi
done <"${act_file}"

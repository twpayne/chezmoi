#!/bin/sh

# chezmoi install script
# contains code from and inspired by
# https://github.com/client9/shlib
# https://github.com/goreleaser/godownloader

set -e

BINDIR="${BINDIR:-$HOME/bin}"
CHEZMOI_USER_REPO="${CHEZMOI_USER_REPO:-twpayne/chezmoi}"
TAGARG=latest
LOG_LEVEL=2

GITHUB_DOWNLOAD="https://github.com/${CHEZMOI_USER_REPO}/releases/download"

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "${tmpdir}"' EXIT
trap 'exit' INT TERM

usage() {
	this="${1}"
	cat <<EOF
${this}: download chezmoi and optionally run chezmoi

Usage: ${this} [-b bindir] [-d] [-t tag] [chezmoi-args]
  -b sets the installation directory, default is ${BINDIR}.
  -d enables debug logging.
  -t sets the tag, default is ${TAG}.
If chezmoi-args are given, after install chezmoi is executed with chezmoi-args.
EOF
	exit 2
}

main() {
	parse_args "${@}"
	shift "$((OPTIND - 1))"

	GOOS="$(get_goos)"
	GOARCH="$(get_goarch)"
	check_goos_goarch "${GOOS}/${GOARCH}"

	TAG="$(real_tag "${TAGARG}")"
	VERSION="${TAG#v}"

	log_info "found version ${VERSION} for ${TAGARG}/${GOOS}/${GOARCH}"

	BINSUFFIX=
	FORMAT=tar.gz
	GOOS_EXTRA=
	case "${GOOS}" in
	linux)
		case "${GOARCH}" in
		amd64)
			case "$(get_libc)" in
			glibc)
				GOOS_EXTRA="-glibc"
				;;
			musl)
				GOOS_EXTRA="-musl"
				;;
			esac
			;;
		esac
		;;
	windows)
		BINSUFFIX=.exe
		FORMAT=zip
		;;
	esac
	case "${GOARCH}" in
	386) arch=i386 ;;
	*) arch="${GOARCH}" ;;
	esac

	# download tarball
	NAME="chezmoi_${VERSION}_${GOOS}${GOOS_EXTRA}_${arch}"
	TARBALL="${NAME}.${FORMAT}"
	TARBALL_URL="${GITHUB_DOWNLOAD}/${TAG}/${TARBALL}"
	http_download "${tmpdir}/${TARBALL}" "${TARBALL_URL}" || exit 1

	# download checksums
	CHECKSUMS="chezmoi_${VERSION}_checksums.txt"
	CHECKSUMS_URL="${GITHUB_DOWNLOAD}/${TAG}/${CHECKSUMS}"
	http_download "${tmpdir}/${CHECKSUMS}" "${CHECKSUMS_URL}" || exit 1

	# verify checksums
	hash_sha256_verify "${tmpdir}/${TARBALL}" "${tmpdir}/${CHECKSUMS}"

	(cd -- "${tmpdir}" && untar "${TARBALL}")

	# install binary
	if [ ! -d "${BINDIR}" ]; then
		install -d "${BINDIR}"
	fi
	BINARY="chezmoi${BINSUFFIX}"
	install -- "${tmpdir}/${BINARY}" "${BINDIR}/"
	log_info "installed ${BINDIR}/${BINARY}"

	if [ -n "${1+n}" ]; then
		exec "${BINDIR}/${BINARY}" "${@}"
	fi
}

parse_args() {
	while getopts "b:dh?t:" arg; do
		case "${arg}" in
		b) BINDIR="${OPTARG}" ;;
		d) LOG_LEVEL=3 ;;
		h | \?) usage "${0}" ;;
		t) TAGARG="${OPTARG}" ;;
		*) return 1 ;;
		esac
	done
}

get_goos() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	case "${os}" in
	cygwin_nt*) goos="windows" ;;
	mingw*) goos="windows" ;;
	msys_nt*) goos="windows" ;;
	*) goos="${os}" ;;
	esac
	printf '%s' "${goos}"
}

get_goarch() {
	arch="$(uname -m)"
	case "${arch}" in
	aarch64) goarch="arm64" ;;
	armv*) goarch="arm" ;;
	i386) goarch="386" ;;
	i686) goarch="386" ;;
	i86pc) goarch="amd64" ;;
	x86) goarch="386" ;;
	x86_64) goarch="amd64" ;;
	*) goarch="${arch}" ;;
	esac
	printf '%s' "${goarch}"
}

check_goos_goarch() {
	case "${1}" in
	darwin/amd64) return 0 ;;
	darwin/arm64) return 0 ;;
	freebsd/386) return 0 ;;
	freebsd/amd64) return 0 ;;
	freebsd/arm) return 0 ;;
	freebsd/arm64) return 0 ;;
	freebsd/riscv64) return 0 ;;
	linux/386) return 0 ;;
	linux/amd64) return 0 ;;
	linux/arm) return 0 ;;
	linux/arm64) return 0 ;;
	linux/loong64) return 0 ;;
	linux/mips64) return 0 ;;
	linux/mips64le) return 0 ;;
	linux/ppc64) return 0 ;;
	linux/ppc64le) return 0 ;;
	linux/riscv64) return 0 ;;
	linux/s390x) return 0 ;;
	openbsd/386) return 0 ;;
	openbsd/amd64) return 0 ;;
	openbsd/arm) return 0 ;;
	openbsd/arm64) return 0 ;;
	windows/386) return 0 ;;
	windows/amd64) return 0 ;;
	windows/arm) return 0 ;;
	*)
		printf '%s: unsupported platform\n' "${1}" 1>&2
		return 1
		;;
	esac
}

get_libc() {
	if is_command ldd; then
		case "$(ldd --version 2>&1 | tr '[:upper:]' '[:lower:]')" in
		*glibc* | *"gnu libc"*)
			printf glibc
			return
			;;
		*musl*)
			printf musl
			return
			;;
		esac
	fi
	if is_command getconf; then
		case "$(getconf GNU_LIBC_VERSION 2>&1)" in
		*glibc*)
			printf glibc
			return
			;;
		esac
	fi
	log_crit "unable to determine libc" 1>&2
	exit 1
}

real_tag() {
	tag="${1}"
	log_debug "checking GitHub for tag ${tag}"
	release_url="https://github.com/${CHEZMOI_USER_REPO}/releases/${tag}"
	json="$(http_get "${release_url}" "Accept: application/json")"
	if [ -z "${json}" ]; then
		log_err "real_tag error retrieving GitHub release ${tag}"
		return 1
	fi
	real_tag="$(printf '%s\n' "${json}" | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//')"
	if [ -z "${real_tag}" ]; then
		log_err "real_tag error determining real tag of GitHub release ${tag}"
		return 1
	fi
	if [ -z "${real_tag}" ]; then
		return 1
	fi
	log_debug "found tag ${real_tag} for ${tag}"
	printf '%s' "${real_tag}"
}

http_get() {
	tmpfile="$(mktemp)"
	http_download "${tmpfile}" "${1}" "${2}" || return 1
	body="$(cat "${tmpfile}")"
	rm -f "${tmpfile}"
	printf '%s\n' "${body}"
}

http_download_curl() {
	local_file="${1}"
	source_url="${2}"
	header="${3}"
	if [ -z "${header}" ]; then
		code="$(curl -w '%{http_code}' -sL -o "${local_file}" "${source_url}")"
	else
		code="$(curl -w '%{http_code}' -sL -H "${header}" -o "${local_file}" "${source_url}")"
	fi
	if [ "${code}" != "200" ]; then
		log_debug "http_download_curl received HTTP status ${code}"
		return 1
	fi
	return 0
}

http_download_wget() {
	local_file="${1}"
	source_url="${2}"
	header="${3}"
	if [ -z "${header}" ]; then
		wget -q -O "${local_file}" "${source_url}" || return 1
	else
		wget -q --header "${header}" -O "${local_file}" "${source_url}" || return 1
	fi
}

http_download() {
	log_debug "http_download ${2}"
	if is_command curl; then
		http_download_curl "${@}" || return 1
		return
	elif is_command wget; then
		http_download_wget "${@}" || return 1
		return
	fi
	log_crit "http_download unable to find wget or curl"
	return 1
}

hash_sha256() {
	target="${1}"
	if is_command sha256sum; then
		hash="$(sha256sum "${target}")" || return 1
		printf '%s' "${hash}" | cut -d ' ' -f 1
	elif is_command shasum; then
		hash="$(shasum -a 256 "${target}" 2>/dev/null)" || return 1
		printf '%s' "${hash}" | cut -d ' ' -f 1
	elif is_command sha256; then
		hash="$(sha256 -q "${target}" 2>/dev/null)" || return 1
		printf '%s' "${hash}" | cut -d ' ' -f 1
	elif is_command openssl; then
		hash="$(openssl dgst -sha256 "${target}")" || return 1
		printf '%s' "${hash}" | cut -d ' ' -f a
	else
		log_crit "hash_sha256 unable to find command to compute SHA256 hash"
		return 1
	fi
}

hash_sha256_verify() {
	target="${1}"
	checksums="${2}"
	basename="${target##*/}"

	want="$(grep "${basename}" "${checksums}" 2>/dev/null | tr '\t' ' ' | cut -d ' ' -f 1)"
	if [ -z "${want}" ]; then
		log_err "hash_sha256_verify unable to find checksum for ${target} in ${checksums}"
		return 1
	fi

	got="$(hash_sha256 "${target}")"
	if [ "${want}" != "${got}" ]; then
		log_err "hash_sha256_verify checksum for ${target} did not verify ${want} vs ${got}"
		return 1
	fi
}

untar() {
	tarball="${1}"
	case "${tarball}" in
	*.tar.gz | *.tgz) tar -xzf "${tarball}" ;;
	*.tar) tar -xf "${tarball}" ;;
	*.zip) unzip -- "${tarball}" ;;
	*)
		log_err "untar unknown archive format for ${tarball}"
		return 1
		;;
	esac
}

is_command() {
	type "${1}" >/dev/null 2>&1
}

log_debug() {
	[ 3 -le "${LOG_LEVEL}" ] || return 0
	printf 'debug %s\n' "${*}" 1>&2
}

log_info() {
	[ 2 -le "${LOG_LEVEL}" ] || return 0
	printf 'info %s\n' "${*}" 1>&2
}

log_err() {
	[ 1 -le "${LOG_LEVEL}" ] || return 0
	printf 'error %s\n' "${*}" 1>&2
}

log_crit() {
	[ 0 -le "${LOG_LEVEL}" ] || return 0
	printf 'critical %s\n' "${*}" 1>&2
}

main "${@}"

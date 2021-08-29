#!/bin/bash

set -eo pipefail

# generate the new website
rm -rf public/
hugo

# clone and checkout the gh-pages branch in a temporary directory
tmpdir=$(mktemp -d)
cleanup() {
    rm -rf "${tmpdir}"
}
trap cleanup EXIT
git branch -f gh-pages origin/gh-pages
git clone --branch=gh-pages --local ../.. "${tmpdir}"

# copy the new website to the temporary directory
rm -rf "${tmpdir:?}"/*
cp -r public/* "${tmpdir}"

# prepare the clone
cd "${tmpdir}"
git checkout CNAME
git remote set-url origin https://github.com/twpayne/chezmoi.git
git fetch origin
git reset origin/gh-pages

# commit the new website
if ! git diff --quiet; then
    git add .
    git commit --message "Update gh-pages"
fi

# give the user the opportunity to push the new website
echo "run git push to push the new website"

${SHELL}

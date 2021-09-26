FROM fedora:latest

RUN dnf update -y && \
    dnf install -y bzip2 git gnupg golang
ENTRYPOINT ( cd /chezmoi && go test ./... )
FROM fedora:latest

ENV GOPROXY=https://proxy.golang.org/

RUN dnf update -y && \
    dnf install -y bzip2 git gnupg golang

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

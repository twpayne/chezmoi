FROM fedora:latest

RUN dnf update -y && \
    dnf install -y bzip2 git gnupg golang

RUN go get golang.org/dl/go1.18 && \
    ${HOME}/go/bin/go1.18 download
ENV GO=/root/go/bin/go1.18

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

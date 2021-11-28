FROM voidlinux/voidlinux:latest

RUN \
    xbps-install --sync --update --yes && \
    xbps-install --yes age gcc git go unzip zip

ENTRYPOINT ( cd /chezmoi && go test ./... )

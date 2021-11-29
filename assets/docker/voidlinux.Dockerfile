FROM voidlinux/voidlinux:latest

RUN \
    xbps-install --sync --update --yes && \
    xbps-install --yes age gcc git go unzip zip

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

FROM ghcr.io/void-linux/void-linux:20220211rc01-full-x86_64

RUN \
    xbps-install --sync --update --yes ; \
    xbps-install --update --yes xbps && \
    xbps-install --sync --update --yes && \
    xbps-install --yes age curl gcc git go unzip zip

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

FROM archlinux:latest

RUN pacman -Syu --noconfirm age gcc git go unzip zip

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

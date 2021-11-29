FROM archlinux:latest

RUN pacman -Sy --noconfirm --noprogressbar age gcc git go unzip zip

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

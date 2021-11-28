FROM archlinux:latest

RUN pacman -Sy --noconfirm --noprogressbar age gcc git go unzip zip

ENTRYPOINT ( cd /chezmoi && go test ./... )

FROM alpine:latest

RUN apk add age git go unzip zip

ENTRYPOINT ( cd /chezmoi && go test ./... )

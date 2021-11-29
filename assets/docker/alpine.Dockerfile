FROM alpine:latest

RUN apk --no-cache add age git go unzip zip

ENTRYPOINT ( cd /chezmoi && go test ./... )

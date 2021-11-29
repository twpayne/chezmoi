FROM alpine:latest

RUN apk --no-cache add age git go unzip zip

COPY assets/docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT /entrypoint.sh

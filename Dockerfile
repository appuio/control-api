FROM docker.io/library/alpine:3.15 as runtime

ENTRYPOINT ["control-api"]

RUN \
  apk add --update --no-cache \
    bash \
    ca-certificates \
    curl

RUN \
  mkdir /.cache && chmod -R g=u /.cache

COPY control-api /usr/local/bin/

RUN chmod a+x /usr/local/bin/control-api

USER 65532:0

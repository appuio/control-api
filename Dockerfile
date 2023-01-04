FROM docker.io/library/alpine:3.17 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

RUN \
  mkdir /.cache && chmod -R g=u /.cache

COPY control-api /usr/local/bin/

RUN chmod a+x /usr/local/bin/control-api

USER 65532:0

ENTRYPOINT ["control-api"]

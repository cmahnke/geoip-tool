FROM golang:1.20 AS builder

ENV MOD_DIR=/go/src/christianmahnke.de/tools/data/geoip-tool

RUN --mount=target=/mnt/build-context \
    mkdir -p $MOD_DIR && \
    cp -r /mnt/build-context/*.go /mnt/build-context/go.* $MOD_DIR && \
    cd $MOD_DIR && \
    CGO_ENABLED=0 go install

FROM alpine:3

LABEL maintainer="cmahnke@gmail.com"
LABEL org.opencontainers.image.source=https://github.com/cmahnke/geoip-tool

COPY --from=builder /go/bin/geoip-tool /usr/local/bin/
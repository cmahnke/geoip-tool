FROM golang:1.20

LABEL maintainer="cmahnke@gmail.com"
LABEL org.opencontainers.image.source=https://github.com/cmahnke/geoip-tool

ENV MOD_DIR=/go/src/christianmahnke.de/tools/data/geoip-tool

RUN --mount=target=/mnt/build-context \
    mkdir -p $MOD_DIR && \
    cp -r /mnt/build-context/*.go /mnt/build-context/go.* $MOD_DIR && \
    cd $MOD_DIR && \
    go install
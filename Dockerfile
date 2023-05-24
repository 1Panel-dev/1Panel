FROM node:18.14 as build_web
ARG TARGETARCH
ARG NPM_REGISTRY="https://registry.npmmirror.com"
ENV NODE_OPTIONS="--max-old-space-size=4096"

WORKDIR /data

RUN set -ex \
    && npm config set registry ${NPM_REGISTRY}

ADD . /data

RUN set -ex \
    && cd /data/frontend \
    && npm install

RUN set -ex \
    && cd /data/frontend \
    && npm run build:dev

FROM golang:1.20
ARG TARGETARCH
ARG GOPROXY="https://goproxy.cn,direct"

COPY --from=build_web /data /data

WORKDIR /data

RUN set -ex \
    && go env -w GOPROXY=${GOPROXY} \
    && go install github.com/goreleaser/goreleaser@latest \
    && goreleaser build --single-target --snapshot --clean

CMD ["/bin/bash"]
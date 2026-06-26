# syntax=docker/dockerfile:1
# 需要 BuildKit（Docker 23+ / Compose v2 默认开启）
# scheduler 无 CGO 依赖（MySQL 纯 Go 驱动），builder 无需 apt 系统包

FROM golang:1.23-bookworm AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

WORKDIR /build

# 先下载依赖，源码变更时复用 mod 缓存层
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/scheduler ./cmd/server


FROM debian:bookworm-slim AS runtime

ARG DEBIAN_MIRROR=mirrors.aliyun.com
ENV TZ=Asia/Shanghai \
    GIN_MODE=release \
    SERVER_ADDR=:9000

RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-scheduler-runtime \
    sed -i "s|deb.debian.org|${DEBIAN_MIRROR}|g; s|security.debian.org|${DEBIAN_MIRROR}|g" /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata wget \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -r -u 10001 -M -d /app app \
    && mkdir -p /app \
    && chown -R app:app /app

WORKDIR /app

COPY --from=builder /out/scheduler /usr/local/bin/scheduler

USER app
EXPOSE 9000

HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:9000/health || exit 1

ENTRYPOINT ["/usr/local/bin/scheduler"]

# syntax=docker/dockerfile:1
# 需要 BuildKit（Docker 23+ / Compose v2 默认开启）
# 构建阶段用 bookworm：官方 golang 镜像已带 gcc/libc6-dev，只需补 sqlite 头文件，比 alpine 装 gcc+musl 快很多
# 生产建议配置远程 MySQL；未配置时回退 SQLite（CGO 链接 libsqlite3）

FROM golang:1.23-bookworm AS builder

ARG DEBIAN_MIRROR=mirrors.aliyun.com
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

WORKDIR /build

# 先下载依赖（无需安装系统包，层缓存更细）
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 仅补 SQLite 开发库（golang 镜像已含 gcc / libc6-dev）
# 注意：勿与 runtime 共用 /var/lib/apt/lists 缓存，并行构建会触发 apt 锁冲突
RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-builder \
    sed -i "s|deb.debian.org|${DEBIAN_MIRROR}|g; s|security.debian.org|${DEBIAN_MIRROR}|g" /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY backend/ ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/server ./cmd/server


FROM debian:bookworm-slim AS runtime

ARG DEBIAN_MIRROR=mirrors.aliyun.com
ENV TZ=Asia/Shanghai \
    GIN_MODE=release \
    SERVER_ADDR=:8000 \
    DB_PATH=/app/data/app.db \
    UPLOAD_DIR=/app/data/uploads

RUN --mount=type=cache,target=/var/cache/apt,id=apt-cache-runtime \
    sed -i "s|deb.debian.org|${DEBIAN_MIRROR}|g; s|security.debian.org|${DEBIAN_MIRROR}|g" /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata wget libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -r -u 10001 -M -d /app app \
    && mkdir -p /app/data/uploads \
    && chown -R app:app /app

WORKDIR /app

COPY --from=builder /out/server /usr/local/bin/server

USER app
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8000/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]

# ---- Build Stage ----
FROM registry.cn-shenzhen.aliyuncs.com/wl4g/golang:1.26-alpine AS builder

WORKDIR /build

# Go module proxy args for GFW environments.
ARG GOPROXY
ARG GONOSUMDB

COPY . .

ARG BUILD_TAGS=
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN if [ -n "${GOPROXY}" ]; then go env -w GOPROXY="${GOPROXY}"; fi && \
    if [ -n "${GONOSUMDB}" ]; then go env -w GONOSUMDB="${GONOSUMDB}"; fi && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags "${BUILD_TAGS}" -trimpath -ldflags="-s -w -X main.versionStr=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}" -o /build/server .

# ---- Runtime Stage ----
# Static Go binary on scratch — no registry pull needed behind GFW.
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/server /usr/local/bin/server

ENV HOME=/home/mcp

EXPOSE 8080 9991

ENTRYPOINT ["/usr/local/bin/server", "--transport", "http", "--port", "8080"]


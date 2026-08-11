# ---- Build Stage ----
FROM registry.cn-shenzhen.aliyuncs.com/wl4g/golang:1.26-alpine AS builder

WORKDIR /build

# Go module proxy args for GFW environments.
ARG GOPROXY
ARG GONOSUMDB
ARG GONOSUMCHECK

COPY . .

ARG BUILD_TAGS=

RUN if [ -n "${GOPROXY}" ]; then go env -w GOPROXY="${GOPROXY}"; fi && \
    if [ -n "${GONOSUMDB}" ]; then go env -w GONOSUMDB="${GONOSUMDB}"; fi && \
    if [ -n "${GONOSUMCHECK}" ]; then go env -w GONOSUMCHECK="${GONOSUMCHECK}"; fi && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags "${BUILD_TAGS}" -trimpath -ldflags="-s -w" -o /build/server .

# ---- Runtime Stage ----
# Static Go binary on scratch — no registry pull needed behind GFW.
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/server /usr/local/bin/server

ENV HOME=/home/mcp

EXPOSE 8080 9991

ENTRYPOINT ["/usr/local/bin/server", "--transport", "http", "--port", "8080"]

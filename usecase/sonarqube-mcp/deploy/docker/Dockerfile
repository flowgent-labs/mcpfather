# ---- Build Stage ----
FROM registry.cn-shenzhen.aliyuncs.com/wl4g/golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

ARG BUILD_TAGS=

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags "${BUILD_TAGS}" -trimpath -ldflags="-s -w" -o /build/server .

# ---- Runtime Stage ----
FROM gcr.io/distroless/static-debian12:nonroot

ENV HOME=/home/mcp

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/server /usr/local/bin/server

EXPOSE 8080 9991

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/server", "--transport", "http", "--port", "8080"]

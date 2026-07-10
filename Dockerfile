# ---- Build Stage ----
FROM registry.cn-shenzhen.aliyuncs.com/wl4g/golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /build/mcpfather ./cmd/mcpfather

# ---- Runtime Stage ----
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/mcpfather /usr/local/bin/mcpfather

ENTRYPOINT ["/usr/local/bin/mcpfather"]

FROM litestream/litestream:latest AS litestream-builder

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o pocketbase ./cmd/pocketbase
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o main ./cmd/pocketstream

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl sqlite
COPY --from=litestream-builder /usr/local/bin/litestream /litestream
COPY --from=builder /app/pocketbase /pocketbase
COPY --from=builder /app/main /main
RUN chmod +x /litestream /pocketbase

RUN mkdir -p /pb_data /pb_backup && chown -R 1000:1000 /pb_data /pb_backup

USER 1000:1000

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8090/api/health || exit 1

EXPOSE 8090

ENTRYPOINT ["/main"]
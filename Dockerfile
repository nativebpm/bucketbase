FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates build-base
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pocketbase ./cmd/pocketbase

FROM litestream/litestream:latest AS litestream-builder

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl sqlite
COPY --from=builder /app/pocketbase /pocketbase
COPY --from=litestream-builder /usr/local/bin/litestream /litestream
RUN chmod +x /litestream

USER 1000:1000

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8090/api/health || exit 1

EXPOSE 8090

CMD ["/pocketbase", "serve", "--http", ":8090"]
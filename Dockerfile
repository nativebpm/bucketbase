FROM litestream/litestream:latest AS litestream-builder

FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates build-base
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pocketbase ./cmd/pocketbase
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/pocketstream

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl sqlite
COPY --from=litestream-builder /usr/local/bin/litestream /litestream
COPY --from=builder /app/pocketbase /pocketbase
COPY --from=builder /app/main /main
RUN chmod +x /litestream
RUN chmod +x /pocketbase

USER 1000:1000

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8090/api/health || exit 1

EXPOSE 8090

ENTRYPOINT ["/main"]
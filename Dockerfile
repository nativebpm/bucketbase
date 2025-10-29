FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates build-base

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/litestream

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pocketbase ./cmd/pocketbase

FROM golang:1.24-alpine AS litestream-builder

RUN go install github.com/benbjohnson/litestream/cmd/litestream@latest

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl sqlite

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000

COPY --from=builder /app/main .
COPY --from=builder /app/pocketbase .
COPY --from=litestream-builder /go/bin/litestream .

RUN mkdir -p pb_backup && chown -R appuser:appgroup /main /pocketbase /litestream /pb_backup

USER appuser

HEALTHCHECK --interval=10s --timeout=10s --start-period=40s --retries=9 \
    CMD curl -f http://localhost:8090/api/health || exit 1

EXPOSE 8090

ENTRYPOINT ["/main"]
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates build-base

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o pocketbase ./cmd/pocketbase

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000

WORKDIR /app

COPY --from=builder /app/pocketbase .

RUN mkdir -p pb_backup

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8090

ENTRYPOINT ["/app/pocketbase"]
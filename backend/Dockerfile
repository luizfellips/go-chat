# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy && go test ./... && CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /server /app/server
COPY migrations /app/migrations
COPY entrypoint.sh /app/entrypoint.sh

RUN sed -i 's/\r$//' /app/entrypoint.sh && chmod +x /app/entrypoint.sh /app/server \
    && chown -R app:app /app

USER app

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --retries=5 --start-period=15s \
  CMD wget -q -O /dev/null http://localhost:8080/ready || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]

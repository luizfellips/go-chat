FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /wsbots .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /wsbots /wsbots

ENTRYPOINT ["/wsbots"]

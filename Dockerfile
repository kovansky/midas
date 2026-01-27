FROM golang:1.18-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ./midasd ./cmd/midasd


FROM alpine:latest

RUN apk --no-cache add curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs
COPY --from=builder /app/midasd /app/midasd

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -fsS http://localhost:8445/system/version || exit 1

EXPOSE 8443
ENTRYPOINT ["/app/midasd"]
CMD ["--config", "/app/config/config.json"]

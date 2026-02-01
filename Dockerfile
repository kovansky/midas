FROM golang:1.18-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ./midasd ./cmd/midasd

FROM node:20.11-alpine AS astro-tools

RUN npm install -g astro@2.2.1

FROM alpine:latest

RUN apk --no-cache add curl libstdc++

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs
COPY --from=builder /app/midasd /app/midasd

COPY --from=astro-tools /usr/local /usr/local

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -fsS http://localhost:8443/system/version || exit 1

EXPOSE 8443
ENTRYPOINT ["/app/midasd"]
CMD ["--config", "/app/config/config.json"]

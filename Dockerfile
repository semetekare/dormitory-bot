FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dormitory-bot ./cmd/dormitory-bot/main.go

FROM debian:stable-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget tzdata && \
    ln -fs /usr/share/zoneinfo/Asia/Irkutsk /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/dormitory-bot .
COPY --from=builder /app/config ./config

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8081/health || exit 1

EXPOSE 8080

CMD ["./dormitory-bot"]

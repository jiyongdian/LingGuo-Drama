FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/spiritfruit ./main.go

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg wget tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/uploads /app/storage/logs

COPY --from=builder /out/spiritfruit /usr/local/bin/spiritfruit
COPY .env.example ./.env
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["spiritfruit", "serve"]

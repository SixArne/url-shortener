# -------- Stage 1: Builder (ARM64) --------
FROM --platform=linux/arm64 golang:1.25.1 AS builder

ENV CGO_ENABLED=1
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

# -------- Stage 2: Runtime (ARM64) --------
FROM --platform=linux/arm64 debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080
CMD ["./app"]

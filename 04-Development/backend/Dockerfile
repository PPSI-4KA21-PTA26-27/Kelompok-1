# --- Build stage ---
FROM golang:1.22-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# --- Run stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S app && adduser -S -G app -u 10001 app
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
COPY migrations /app/migrations

# Tidak ada secret di image: semua konfigurasi datang dari environment saat runtime.
USER app
ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" >/dev/null || exit 1

CMD ["/app/api"]

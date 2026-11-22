# Stage 1: Compile the Go application.
FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# Stage 2: Run the compiled application.
FROM alpine:3.23

WORKDIR /app

# Run as a non-root user.
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder --chown=app:app /out/server ./server
COPY --chown=app:app web/ ./web/

USER app

ENV LISTEN_ADDR=:8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1

CMD ["./server"]
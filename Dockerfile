# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Dependencies are cached separately from code changes
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY . .

# Run test suite during build to prevent deploying broken code
RUN go test ./...

# Build statically linked binary without debug symbols or host path leaks
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /app/bin/api ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates for outbound HTTPS (AI provider / OAuth); tzdata for streak calculation
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/bin/api /app/api

# Non-root unprivileged execution
USER nobody:nobody

EXPOSE 8080

# Native container healthcheck
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/app/api"]

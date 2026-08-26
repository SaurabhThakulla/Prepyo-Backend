# ---- build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Dependencies are copied first so a source-only change does not re-download
# the module cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off produces a static binary that runs on a bare image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

RUN go test ./...

# ---- run ----
FROM alpine:3.20

# ca-certificates for outbound HTTPS to the AI provider; tzdata because streak
# day boundaries are computed in the learner's own timezone.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/api /app/api

USER nobody:nobody
EXPOSE 8080

ENTRYPOINT ["/app/api"]

# ======================
# Stage 1: Build
# ======================
FROM golang:1.25-alpine AS builder

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o sso-server ./cmd/server

# ======================
# Stage 2: Runtime
# ======================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/internal/db/migrations .internal/db/migrations

COPY --from=builder /app/sso-server /app/sso-server

RUN adduser -D appuser
USER appuser

EXPOSE 9000

# ✅ THIS IS THE FIX
ENTRYPOINT ["/app/sso-server"]

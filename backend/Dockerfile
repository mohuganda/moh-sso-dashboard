# ======================
# Stage 1: Build
# ======================
FROM --platform=$BUILDPLATFORM golang:1.25.4-alpine AS builder

# BuildKit supplies these from --platform. Do not provide defaults here: a
# default TARGETARCH=amd64 silently creates an amd64 binary in an arm64 image.
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown
ARG DIRTY=false

ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build \
  -trimpath \
  -ldflags="-s -w \
  -X 'github.com/moh-sso-dashboard/internal/version.Version=${VERSION}' \
  -X 'github.com/moh-sso-dashboard/internal/version.Commit=${COMMIT}' \
  -X 'github.com/moh-sso-dashboard/internal/version.BuildTime=${BUILD_TIME}' \
  -X 'github.com/moh-sso-dashboard/internal/version.Dirty=${DIRTY}'" \
  -o sso-server ./cmd/server

# ======================
# Stage 2: Runtime
# ======================
FROM alpine:3.19

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown
ARG SOURCE_URL=https://github.com/mohuganda/moh-sso-dashboard

LABEL org.opencontainers.image.title="MOH SSO Dashboard Backend" \
  org.opencontainers.image.version="$VERSION" \
  org.opencontainers.image.revision="$COMMIT" \
  org.opencontainers.image.created="$BUILD_TIME" \
  org.opencontainers.image.source="$SOURCE_URL"

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
COPY --from=builder /app/config ./config
COPY --from=builder /app/sso-server ./sso-server

# Fail the image build if the runtime migration set is incomplete.
RUN test -f ./internal/db/migrations/000001_initial_schema.up.sql \
  && test -f ./internal/db/migrations/000010_create_schema_facilities.up.sql \
  && test -f ./internal/db/migrations/000039_add_announcement_group_audience.up.sql

RUN adduser -D appuser
USER appuser

EXPOSE 9000

ENTRYPOINT ["./sso-server"]

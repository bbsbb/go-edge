# Shared multi-stage Dockerfile for building application modules.
# Usage: docker build -f package/Dockerfile --build-arg APP=<app> -t go-edge/<app> .

ARG GO_VERSION=1.25.7

# --- Build stage ---
FROM golang:${GO_VERSION}-alpine AS builder

ARG APP

WORKDIR /src

# Copy go.mod/go.sum for dependency caching.
# Core must be copied first since app modules use replace directives.
COPY core/go.mod core/go.sum ./core/
COPY apps/${APP}/go.mod apps/${APP}/go.sum ./apps/${APP}/
RUN cd apps/${APP} && go mod download

# Copy source code.
COPY core/ ./core/
COPY apps/${APP}/ ./apps/${APP}/

# Build static binary.
RUN cd apps/${APP} && CGO_ENABLED=0 go build -o /app/${APP} .

# --- Runtime stage ---
FROM alpine:3.21

ARG APP

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/${APP} ./${APP}
COPY apps/${APP}/resources/ ./resources/
COPY package/${APP}/entrypoint.sh ./entrypoint.sh

USER app

ENTRYPOINT ["/app/entrypoint.sh"]

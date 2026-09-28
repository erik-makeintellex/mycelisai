# Build Stage
FROM golang:1.26 AS builder

WORKDIR /app

# Copy dependency files
COPY core/go.mod core/go.sum ./
COPY core/pkg ./core/pkg
COPY proto ./proto

# We need to ensure the local module directives work.
# go.mod points to github.com/mycelis/core
# We assume the build context is the repo root.

# Download deps. proxy.golang.org intermittently resets parallel module
# fetches ("read tcp ...: connection reset by peer"); a BuildKit cache mount
# keeps already-fetched modules on disk between attempts and across builds,
# a bounded retry loop with backoff re-requests only what is still missing,
# and GOMAXPROCS=2 lowers fetch parallelism (`go mod download` sizes its fetch
# pool from GOMAXPROCS; the -p build flag does not apply to it), which reduces
# how often the proxy resets connections. A real, non-transient failure still
# exits non-zero.
RUN --mount=type=cache,target=/go/pkg/mod \
    export GOMAXPROCS=2; \
    attempt=1; max_attempts=5; \
    until go mod download; do \
      if [ "$attempt" -ge "$max_attempts" ]; then \
        echo "go mod download failed after $max_attempts attempts" >&2; \
        exit 1; \
      fi; \
      echo "go mod download attempt $attempt failed, retrying in $((attempt * 5))s..." >&2; \
      sleep $((attempt * 5)); \
      attempt=$((attempt + 1)); \
    done

# Copy source
COPY core/ ./core/

# Build
# CGO_ENABLED=0 for static binary
# We switch to core directory to build within the module context
WORKDIR /app/core
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Run Stage
# Curated stdio MCP servers are launched through npm/npx, so the self-hosted
# runtime image carries Node's package runner alongside the Go server.
FROM node:22-bookworm-slim

WORKDIR /core

COPY core/output-validation/package*.json /core/output-validation/
RUN cd /core/output-validation && npm ci && PLAYWRIGHT_DOWNLOAD_CONNECTION_TIMEOUT=120000 npx playwright install --with-deps chromium
COPY interface/scripts/validate-generated-output.mjs /core/output-validation/validate-generated-output.mjs
COPY --from=builder /server /server
COPY --from=builder /app/core/config /core/config
COPY cognitive/config/engine.yaml /core/config/engine.yaml

ENV MYCELIS_OUTPUT_VALIDATOR_SCRIPT=/core/output-validation/validate-generated-output.mjs
ENV MYCELIS_OUTPUT_VALIDATOR_WORKDIR=/core/output-validation

ENTRYPOINT ["/server"]

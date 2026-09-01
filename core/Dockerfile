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

# Download deps
RUN go mod download

# Copy source
COPY core/ ./core/

# Build
# CGO_ENABLED=0 for static binary
# We switch to core directory to build within the module context
WORKDIR /app/core
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Run Stage
# Curated stdio MCP servers are launched through npm/npx, so the self-hosted
# runtime image carries Node's package runner alongside the Go server.
FROM node:22-bookworm-slim

WORKDIR /core

COPY --from=builder /server /server
COPY --from=builder /app/core/config /core/config
COPY cognitive/config/engine.yaml /core/config/engine.yaml
COPY core/output-validation/package*.json /core/output-validation/
RUN cd /core/output-validation && npm ci && npx playwright install --with-deps chromium
COPY interface/scripts/validate-generated-output.mjs /core/output-validation/validate-generated-output.mjs

ENV MYCELIS_OUTPUT_VALIDATOR_SCRIPT=/core/output-validation/validate-generated-output.mjs
ENV MYCELIS_OUTPUT_VALIDATOR_WORKDIR=/core/output-validation

ENTRYPOINT ["/server"]

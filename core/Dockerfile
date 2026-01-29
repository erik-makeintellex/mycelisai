# Build Stage
FROM golang:1.24 AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

# Run Stage
FROM gcr.io/distroless/base-debian12

COPY --from=builder /server /server

ENTRYPOINT ["/server"]

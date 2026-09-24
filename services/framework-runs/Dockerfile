FROM golang:1.26.0 AS build

WORKDIR /src
COPY services/framework-runs/go.mod services/framework-runs/go.sum ./
RUN go mod download
COPY services/framework-runs/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /framework-runs ./cmd/framework-runs

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /framework-runs /framework-runs
USER 1000:1000
ENTRYPOINT ["/framework-runs"]

FROM golang:1.24-alpine AS builder
WORKDIR /app

# First copy the Go module files
COPY go.mod go.sum ./
RUN go mod download

# Now copy the actual source files
COPY cmd/worker cmd/worker
COPY pkg pkg

# Build the binary from within the module
RUN go build -o worker ./cmd/worker

FROM alpine:latest
WORKDIR /app

# Optional: tini for better signal handling
RUN apk add --no-cache ca-certificates tini

COPY --from=builder /app/worker /app/worker
COPY cluster/worker/worker-entrypoint.sh /usr/local/bin/worker-entrypoint.sh
RUN chmod +x /usr/local/bin/worker-entrypoint.sh

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/worker-entrypoint.sh"]


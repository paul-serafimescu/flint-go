FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files and generated protos
COPY cmd/driver cmd/driver
COPY pkg pkg

# Build the driver binary
RUN go build -o driver ./cmd/driver

# Final minimal image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/driver /app/driver
ENTRYPOINT ["/app/driver"]

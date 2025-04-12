FROM golang:1.24-alpine AS builder
WORKDIR /app

# First copy the Go module files
COPY go.mod go.sum ./
RUN go mod download

# Now copy the actual source files
COPY cmd/driver cmd/driver
COPY pkg pkg

# Build the binary from within the module
RUN go build -o driver ./cmd/driver

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/driver /app/driver
ENTRYPOINT ["/app/driver"]

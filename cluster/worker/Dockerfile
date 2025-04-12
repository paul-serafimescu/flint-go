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
COPY --from=builder /app/worker /app/worker
ENTRYPOINT ["/app/worker"]

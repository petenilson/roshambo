# Build stage
FROM golang:1.23-bullseye AS builder

# Set environment variables for x86_64 build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application for x86_64
RUN go build -o main ./cmd/

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/main .

# Set entrypoint
ENTRYPOINT ["./main"]

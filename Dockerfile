# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod and sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Build with optimizations (static binary, no debug info)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api .

# Stage 2: Run the binary in a minimal image
FROM alpine:3.19

WORKDIR /app

# Add CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy the built binary from builder
COPY --from=builder /app/api .

# Expose API port (update if your app uses another)
EXPOSE 8080

# Run the binary
CMD ["./api"]

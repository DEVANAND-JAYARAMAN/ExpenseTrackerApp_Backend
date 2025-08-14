# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Build with optimizations (static binary, no debug info)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o api .

# Stage 2: Run the binary in a minimal image
FROM alpine:3.19

WORKDIR /app

# Install CA certificates and netcat for wait script
RUN apk add --no-cache ca-certificates netcat-openbsd

# Copy the compiled binary from the builder stage
COPY --from=builder /app/api .

# Copy wait script
COPY wait-for-db.sh .
RUN chmod +x wait-for-db.sh

# Expose API port
EXPOSE 8080

# Wait for DB and then start API
ENTRYPOINT ["sh", "-c", "./wait-for-db.sh \"$DB_HOST\" \"$DB_PORT\" && ./api"]

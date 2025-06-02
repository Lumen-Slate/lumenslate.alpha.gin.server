# Stage 1: Build
FROM golang:1.23-bullseye AS builder

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire app
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: Minimal runtime
FROM alpine:latest

# Install CA certs and curl (for healthcheck)
RUN apk add --no-cache ca-certificates curl

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Port Cloud Run expects
EXPOSE 8080

# Optional: Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl --fail http://localhost:8080/health || exit 1

# Start the app
CMD ["./server"]

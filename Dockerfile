FROM golang:1.24.4-alpine

# Install necessary dependencies
RUN apk add --no-cache git curl

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source code
COPY . .

# Expose the port your app uses
EXPOSE 8080

# Health check (optional)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl --fail http://localhost:8080/health || exit 1

# Run the app
CMD ["go", "run", "main.go"]

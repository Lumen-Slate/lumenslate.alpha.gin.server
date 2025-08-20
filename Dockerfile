FROM golang:1.24.4-alpine

# Install necessary dependencies
RUN apk add --no-cache git curl ca-certificates openssl

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source code
COPY . .

# Expose the port your app uses
EXPOSE 8080

# Run the app
CMD ["go", "run", "main.go"]

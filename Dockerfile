FROM golang:1.24.4-alpine

# Install necessary dependencies
RUN apk add --no-cache git curl ca-certificates openssl

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy all Go source files from the root directory
COPY . ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /lumenslate-server

EXPOSE 8080

CMD ["/lumenslate-server"]
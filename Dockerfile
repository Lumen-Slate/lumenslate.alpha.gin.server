FROM golang:1.25.0-trixie

# Install necessary dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl ca-certificates openssl \
 && rm -rf /var/lib/apt/lists/*

RUN apt-get install -y pkg-config python3-dev default-libmysqlclient-dev build-essential
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
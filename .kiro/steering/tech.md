# Technology Stack

## Core Technologies

- **Language**: Go 1.24.2
- **Web Framework**: Gin (HTTP router and middleware)
- **Database**: MongoDB with official Go driver
- **API Documentation**: Swagger/OpenAPI with Swaggo
- **Communication**: gRPC with Protocol Buffers
- **Environment**: Docker containerization support

## Key Libraries

- `gin-gonic/gin` - HTTP web framework
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `swaggo/gin-swagger` - Swagger documentation
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers
- `joho/godotenv` - Environment variable management
- `google/uuid` - UUID generation
- `gin-contrib/cors` - CORS middleware

## Build System & Commands

### Development Commands
```bash
make run         # Run the server locally
make dev         # Run with Swagger docs regeneration
make dev-all     # Run with both Swagger and Proto regeneration
```

### Build Commands
```bash
make build       # Build production binary
make swagger     # Generate Swagger documentation
make proto       # Generate protobuf Go code from .proto files
```

### Maintenance Commands
```bash
make clean       # Remove generated files and binaries
make fmt         # Format Go code
make test        # Run tests
go mod tidy      # Clean up dependencies
```

### Docker Commands
```bash
docker-compose up    # Run with Docker Compose
```

## Environment Configuration

- Uses `.env` file for local development
- Supports `/secrets/ENV_FILE` for containerized deployments
- Key variables: `PORT`, `MONGO_URI`
- Default port: 8080

## API Documentation

- Swagger UI available at `/docs/index.html`
- Auto-generated from code annotations
- Postman collections provided for testing
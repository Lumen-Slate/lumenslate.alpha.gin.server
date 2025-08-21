# Project settings
APP_NAME = server
CMD_DIR = .
SWAG_OUTPUT = internal/docs
PROTO_SRC = internal/proto/ai_service.proto
PROTO_OUT = internal/proto/ai_service

# Build the application binary
build:
	go build -o $(APP_NAME) $(CMD_DIR)

# Run the application
run:
	go run $(CMD_DIR)/main.go

# Generate Swagger docs
swagger:
	swag init --output $(SWAG_OUTPUT) --parseDependency --parseInternal

# Generate protobuf Go code
proto:
	protoc --go_out=. --go-grpc_out=. internal/proto/ai_service.proto

# Run with Swagger and Proto regeneration
dev-all:
	make swagger
	make proto
	go run $(CMD_DIR)/main.go

# Run with Swagger regeneration
dev:
	make swagger
	go run $(CMD_DIR)/main.go

# Clean generated files
clean:
	rm -f $(APP_NAME)
	rm -rf $(SWAG_OUTPUT)/*
	rm -f $(PROTO_OUT)/*.pb.go

# Format the code
fmt:
	go fmt ./...

# Run tests (if any)
test:
	go test ./...

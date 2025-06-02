# Project settings
APP_NAME = server
CMD_DIR = cmd/server
SWAG_OUTPUT = internal/docs

# Build the application binary
build:
	go build -o $(APP_NAME) $(CMD_DIR)

# Run the application
run:
	go run $(CMD_DIR)/main.go

# Generate Swagger docs
swagger:
	swag init --output $(SWAG_OUTPUT) --parseDependency --parseInternal

# Run with Swagger regeneration
dev:
	make swagger
	go run $(CMD_DIR)/main.go

# Clean generated files
clean:
	rm -f $(APP_NAME)
	rm -rf $(SWAG_OUTPUT)/*

# Format the code
fmt:
	go fmt ./...

# Run tests (if any)
test:
	go test ./...

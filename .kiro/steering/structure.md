# Project Structure & Architecture

## Directory Organization

```
lumenslate/
├── main.go                     # Application entry point
├── internal/                   # Private application code
│   ├── controller/             # HTTP request handlers
│   │   ├── questions/          # Question-specific controllers
│   │   └── *_controller.go     # Entity controllers
│   ├── model/                  # Data models and structs
│   │   ├── questions/          # Question type models
│   │   └── *.go               # Entity models
│   ├── routes/                 # Route definitions and grouping
│   │   ├── questions/          # Question route handlers
│   │   └── *.go               # Entity route files
│   ├── repository/             # Data access layer
│   ├── service/                # Business logic layer
│   ├── serializer/             # Request/response serialization
│   ├── db/                     # Database connection logic
│   ├── grpc_service/           # gRPC client implementations
│   ├── proto/                  # Protocol buffer definitions
│   ├── docs/                   # Auto-generated Swagger docs
│   ├── utils/                  # Utility functions
│   ├── asynq/                  # Async task processing
│   └── machinery/              # Task queue machinery
├── templates/                  # HTML templates
├── tasks/                      # Background task definitions
└── *.json                     # Postman collections
```

## Architecture Patterns

### MVC-Style Structure
- **Controllers**: Handle HTTP requests, validation, and responses
- **Models**: Define data structures and business entities
- **Routes**: Group and organize API endpoints
- **Repository**: Abstract data access operations
- **Service**: Contain business logic and orchestration

### Naming Conventions
- Files: `snake_case` with descriptive suffixes (`*_controller.go`, `*_model.go`)
- Packages: lowercase, single word when possible
- Functions: PascalCase for exported, camelCase for private
- Structs: PascalCase with clear, descriptive names

### Code Organization
- Group related functionality in subdirectories (e.g., `questions/`)
- Separate concerns: controllers handle HTTP, services handle business logic
- Use dependency injection pattern for database and external services
- Keep models focused on data structure, avoid business logic

### API Structure
- RESTful endpoints following standard conventions
- Swagger annotations on all public endpoints
- Consistent error handling and response formats
- Route grouping by entity type

### Question Types Architecture
Special handling for educational question types:
- MCQ (Multiple Choice Questions)
- MSQ (Multiple Select Questions) 
- NAT (Numerical Answer Type)
- Subjective (Open-ended questions)

Each question type has dedicated:
- Model definitions in `internal/model/questions/`
- Controllers in `internal/controller/questions/`
- Routes in `internal/routes/questions/`
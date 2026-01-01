# Productivity Timer

A productivity timer application built with Go, Gin, HTMX, Alpine.js, Templ, and MongoDB with OAuth authentication.

## Features

- Start/stop/reset timer sessions with custom tags
- Track time spent on various tasks
- View statistics and summaries by time period
- OAuth authentication (Google, GitHub, etc.)

## Tech Stack

- **Backend:** Go with Gin framework
- **Frontend:** HTMX + Alpine.js + Templ templates
- **Database:** MongoDB
- **Authentication:** OAuth via Goth

## Project Structure

```
productivity-timer/
├── cmd/api/              # Application entrypoint
├── docs/                 # Generated Swagger/OpenAPI documentation
├── internal/
│   ├── auth/             # Authentication logic
│   ├── database/         # Database operations
│   ├── models/           # Data models
│   └── server/           # HTTP handlers and routing
└── web/templates/        # Templ templates
```

## API Documentation

Once the server is running, access the Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

### API Endpoints

Timer and Stats API endpoints are versioned under `/api/v1`. Auth routes remain at root level for OAuth provider compatibility.

#### Auth Routes (Root Level)

| Method | Endpoint                   | Description    |
| ------ | -------------------------- | -------------- |
| GET    | `/auth/:provider`          | Initiate OAuth |
| GET    | `/auth/:provider/callback` | OAuth callback |
| GET    | `/logout/:provider`        | Logout         |

#### API v1 Routes

| Method | Endpoint                          | Description             |
| ------ | --------------------------------- | ----------------------- |
| GET    | `/health`                         | Health check            |
| POST   | `/api/v1/timer/start`             | Start timer             |
| POST   | `/api/v1/timer/stop`              | Stop timer              |
| POST   | `/api/v1/timer/reset`             | Reset/complete timer    |
| GET    | `/api/v1/stats/summary`           | Get stats summary       |
| GET    | `/api/v1/stats/tag/:tag/sessions` | Get tag sessions        |
| DELETE | `/api/v1/stats/tag/:tag`          | Delete tag and sessions |

## Development

### Prerequisites

- Go 1.21+
- MongoDB
- [swag](https://github.com/swaggo/swag) for Swagger generation

### Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure environment variables
3. Install dependencies: `go mod download`
4. Install swag: `make swagger-install`

### Running

```bash
# Build and run
make run

# Development with hot reload
make watch

# Generate Swagger docs
make swagger
```

### Makefile Commands

- `make build` - Build the application (includes swagger generation)
- `make run` - Build and run
- `make watch` - Hot reload development
- `make swagger` - Generate Swagger documentation
- `make swagger-install` - Install swag CLI tool
- `make clean` - Clean build artifacts

## License

MIT

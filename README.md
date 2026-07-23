# User CRUD API (Go)

Production-ready Go API with Fiber, PostgreSQL, Redis, and JWT authentication.

## Tech Stack

- **Go 1.22** - Language
- **Fiber** - Web framework
- **PostgreSQL** - Database
- **Redis** - Caching
- **JWT** - Authentication
- **Docker Compose** - Containerization

## Quick Start

```bash
# Start infrastructure
docker compose up -d postgres redis

# Install dependencies
go mod tidy

# Run
go run ./cmd/api
```

## Docker

```bash
docker compose up --build
```
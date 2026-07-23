# User CRUD API

Production-ready REST API with PostgreSQL, Redis, and JWT authentication.

## Tech Stack

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

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Register new user |
| `POST` | `/api/v1/auth/login` | Login and get tokens |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users` | List all users |
| `GET` | `/api/v1/users/{id}` | Get user by ID |
| `POST` | `/api/v1/users` | Create user |
| `PATCH` | `/api/v1/users/{id}` | Update user |
| `DELETE` | `/api/v1/users/{id}` | Delete user |

### Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/tasks` | List tasks |
| `GET` | `/api/v1/tasks/{id}` | Get task by ID |
| `POST` | `/api/v1/tasks` | Create task |
| `PATCH` | `/api/v1/tasks/{id}` | Update task |
| `DELETE` | `/api/v1/tasks/{id}` | Delete task |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP__DATABASE_URL` | postgres://... | PostgreSQL URL |
| `APP__REDIS_URL` | redis://... | Redis URL |
| `APP__JWT_SECRET` | (required) | JWT signing secret |
| `APP__JWT_EXPIRATION_HOURS` | 24 | Token expiry |
| `APP__JWT_REFRESH_EXPIRATION_DAYS` | 7 | Refresh token expiry |

## License

MIT
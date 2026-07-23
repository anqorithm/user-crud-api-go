# User CRUD API

[![Go](https://img.shields.io/badge/Go-1.22-blue)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-green)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-red)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-yellow)](https://opensource.org/licenses/MIT)

A production-ready Go API with PostgreSQL, Redis, and JWT authentication.

## Features

- **Fiber** - Fast, minimal web framework
- **PostgreSQL** - Robust relational database
- **Redis** - Caching and session management
- **pgx** - High-performance database driver
- **JWT Authentication** - Secure token-based auth with refresh tokens
- **Docker Compose** - Containerized development and deployment
- **API Versioning** - `/api/v1/` prefix for clean API evolution

## Architecture Diagrams

### System Architecture

```mermaid
flowchart TB
    subgraph Clients["Internet Clients"]
        Browser["Web Browser"]
        MobileApp["Mobile App"]
        ApiClients["API Clients"]
    end

    subgraph Gateway["API Gateway"]
        TLS["TLS Termination"]
    end

    subgraph GoAPI["Go API (Fiber)"]
        MW["Middleware<br/>CORS, Logging"]
        
        subgraph Routes["API Routes"]
            Auth["/auth/*"]
            Users["/users/*"]
            Tasks["/tasks/*"]
        end
        
        subgraph Handlers["Handlers"]
            AuthH["AuthHandler"]
            UserH["UserHandler"]
            TaskH["TaskHandler"]
        end
        
        MW --> Routes --> Handlers
    end

    subgraph DataStore["Data Stores"]
        PG[("PostgreSQL")]
        RD[("Redis")]
    end

    Clients --> TLS --> MW
    Handlers --> PG
    Handlers --> RD
```

### Request Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant API as Fiber
    participant H as Handlers
    participant DB as Database
    participant Cache as Redis

    C->>API: POST /api/v1/auth/register
    API->>H: user_data
    H->>H: Hash password with bcrypt
    H->>DB: INSERT user
    DB-->>H: user created
    H->>Cache: invalidate("users:list:*")
    Cache-->>H: OK
    H->>API: Token response
    API-->>C: 201 Created + JWT token

    Note over C: Use token for protected routes

    C->>API: GET /api/v1/tasks
    API->>H: token
    H->>Cache: get("tasks:user:*")
    
    alt Cache Hit
        Cache-->>H: cached data
    else Cache Miss
        H->>DB: SELECT tasks
        DB-->>H: task list
        H->>Cache: setex(..., ttl)
    end

    H-->>API: task list
    API-->>C: 200 OK
```

### Database Schema

```mermaid
erDiagram
    USERS {
        uuid id PK "Primary Key"
        string name "Required"
        string email UK "Unique"
        int age "Optional"
        bool is_active "Default: true"
        bool is_superuser "Default: false"
        string hashed_password "Required"
        timestamp created_at "Auto"
        timestamp updated_at "Auto"
    }

    TASKS {
        uuid id PK "Primary Key"
        string title "Required"
        string description "Optional"
        int priority "1-5, Default: 1"
        bool completed "Default: false"
        uuid user_id FK "References Users"
        timestamp created_at "Auto"
        timestamp updated_at "Auto"
    }

    USERS ||--o{ TASKS : "creates"
```

### Caching Strategy

```mermaid
flowchart LR
    subgraph ReadOps["Read Operations"]
        A["GET /users<br/>/tasks/{id}"] --> B{Redis Cache}
        B -->|HIT| C["Return Cached"]
        B -->|MISS| D["Query PostgreSQL"]
        D --> E["Store in Redis<br/>TTL: 300s"]
        E --> C
    end

    subgraph WriteOps["Write Operations"]
        F["POST/PATCH/DELETE"] --> G["Write to PostgreSQL"]
        G --> H["Invalidate Cache"]
        H --> I["Return Response"]
    end
```

### Docker Architecture

```mermaid
graph TB
    subgraph Compose["Docker Compose"]
        subgraph Services["Services"]
            API["Go API<br/>:8000"]
            PG["PostgreSQL<br/>:5432"]
            RD["Redis<br/>:6379"]
        end
        
        API --> PG
        API --> RD
    end

    subgraph Clients["Clients"]
        Browser["Browser"]
        Mobile["Mobile App"]
        Client["API Clients"]
    end

    Browser --> API
    Mobile --> API
    Client --> API
```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose

### Local Development

```bash
# Clone the repository
git clone https://github.com/anqorithm/user-crud-api-go
cd user-crud-api-go

# Install dependencies
go mod tidy

# Start infrastructure with Docker
docker compose up -d postgres redis

# Run
go run ./cmd/api
```

The API will be available at:
- **API**: http://localhost:8000

### Using Docker (Full Stack)

```bash
# Build and start all services
docker compose up --build

# Stop all services
docker compose down
```

## Project Structure

```
cmd/
└── api/
    └── main.go              # Application entry point
internal/
├── config/
│   └── config.go           # Configuration
├── handlers/
│   └── handlers.go         # HTTP handlers
├── models/
│   └── models.go          # Data models
└── middleware/            # JWT middleware
pkg/                        # Shared packages
├── redis/                 # Redis client
└── database/              # Database utilities
```

## Configuration

All configuration is managed through environment variables.

```env
APP_NAME=User CRUD API
DATABASE_URL=postgres://postgres:postgres@localhost:5432/users_db
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24
JWT_REFRESH_EXPIRATION_DAYS=7
```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Register new user |
| `POST` | `/api/v1/auth/login` | Login and get tokens |
| `POST` | `/api/v1/auth/refresh` | Refresh token |

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

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |

## License

MIT License - see LICENSE file for details.

---

**Built with Go, PostgreSQL, Redis**
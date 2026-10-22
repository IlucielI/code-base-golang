# Go Core API Boilerplate (`code-base-golang`)

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Clean Architecture](https://img.shields.io/badge/Architecture-Clean%20Architecture-blueviolet?style=flat)]()
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

Production-grade, clean architecture Go REST API boilerplate built with Gin, GORM, PostgreSQL, Redis, RabbitMQ, S3/MinIO, and SMTP/Mailpit.

---

## Table of Contents

- [Architecture & Design Principles](#architecture--design-principles)
- [Directory Structure](#directory-structure)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Running the Application](#running-the-application)
- [Database Migrations & Auto-Migration](#database-migrations--auto-migration)
- [Interactive API Documentation](#interactive-api-documentation)
- [Built-in Infrastructure & Middlewares](#built-in-infrastructure--middlewares)
- [Makefile Commands](#makefile-commands)
- [Running Tests](#running-tests)
- [Core Endpoints](#core-endpoints)
- [Deployment](#deployment)

---

## Architecture & Design Principles

The boilerplate adheres strictly to **Clean Architecture** and idiomatic Go design patterns:

```
                      [ External World ]
            (HTTP Clients, AMQP Queues, Webhooks)
                              │
                              ▼
            ┌───────────────────────────────────┐
            │         Delivery Layer            │
            │  (Gin Controllers, Async Workers) │
            └─────────────────┬─────────────────┘
                              │
                              ▼
            ┌───────────────────────────────────┐
            │          Service Layer            │
            │      (Pure Business Logic)        │
            │   Depends only on Port Interfaces │
            └─────────┬─────────────────┬───────┘
                      │                 │
         ┌────────────▼──────┐   ┌──────▼────────────┐
         │ Repository Ports  │   │ Adapter Ports     │
         │  (GORM / SQL)     │   │ FileStorage       │
         │                   │   │ EventPublisher    │
         │                   │   │ EventSubscriber   │
         │                   │   │ EmailSender       │
         └───────────────────┘   └───────────────────┘
```

1. **Separation of Concerns**: HTTP delivery, domain orchestration, and infrastructure adapters are strictly isolated.
2. **Decoupled Ports & Adapters**: Core services interact solely through Go-idiomatic consumer interfaces (`FileStorage`, `EventPublisher`, `EventSubscriber`, `EmailSender`). Concrete implementations (S3, RabbitMQ, SMTP) can be swapped or mocked with zero business logic changes.
3. **Dependency Injection**: Controllers, services, and repositories receive dependencies via constructors without relying on global state.
4. **Configuration as Code**: Strongly-typed configuration parsed cleanly from environment variables with sensible defaults.
5. **Declarative Routing**: Routes are declared in `internal/routes/routes.yaml` and resolved cleanly at startup.

---

## Directory Structure

```
.
├── cmd/
│   └── api/                # Single unified binary entrypoint (HTTP Server & Migration CLI)
├── deployment/             # Multi-stage Dockerfile, docker-compose, and build scripts
├── docs/                   # Interactive Scalar API docs & OpenAPI 3.0 specification
├── migrations/             # Versioned SQL migration files (.up.sql / .down.sql)
└── internal/
    ├── adapters/           # Infrastructure adapter implementations
    │   ├── database/       # PostgreSQL / GORM connection pool
    │   ├── rabbitmq/       # RabbitMQ AMQP message broker & DLQ adapter
    │   ├── redis/          # Redis client wrapper with JSON helpers
    │   ├── s3/             # MinIO / AWS S3 storage adapter
    │   └── smtp/           # SMTP email dispatcher adapter
    ├── config/             # Environment variable parsing and validation
    ├── constants/          # Application-wide error codes, statuses, token types, user roles, log levels, and response constants
    ├── controllers/        # HTTP handlers parsing requests and rendering JSON responses
    ├── dtos/               # Request payloads and API response definitions (*_dto.go)
    ├── middlewares/        # Middlewares (CORS, Rate Limiter, Auth, Anon Auth, Optional Auth, Basic Auth, Permissions, Client Meta, Logger, Recovery, Body Limit)
    ├── models/             # Domain entities and base models (*_model.go, e.g. BaseModel, JSONMap)
    ├── pkg/                # Reusable packages (AppError, Context Metadata, JWT utilities, Migration, String utilities `strutil`)
    ├── queue/              # Queue message payloads (Id, MustParse, domain payloads)
    ├── repositories/       # Data access layer interfacing with GORM & Redis
    ├── routes/             # YAML-based route registration (routes.yaml) and engine setup
    ├── services/           # Core business logic orchestration (decoupled ports)
    ├── validations/        # Struct input validations using Ozzo-Validation (*_validation.go)
    └── workers/            # Background queue subscribers and worker pool orchestration
```

### File Naming Conventions

To keep each layer's files self-describing, files in these packages carry a mandatory role suffix (applied to every file in the package, shared/base files included):

| Package | Suffix | Examples |
|---------|--------|----------|
| `internal/dtos/` | `*_dto.go` | `recording_dto.go`, `response_dto.go`, `pagination_dto.go` |
| `internal/models/` | `*_model.go` | `recording_model.go`, `base_model.go` |
| `internal/validations/` | `*_validation.go` | `recording_validation.go` |

Test files append the suffix before `_test.go` (e.g. `recording_dto_test.go`).

Other packages name files by topic rather than role — e.g. `internal/constants/` uses one file per topic/domain (`error.go`, `response.go`, `worker.go`, `auth.go`, `logger.go`, `<domain>.go`), not `*_constant.go`. Magic strings and numbers for token types, user roles, and log levels are defined in `constants/`.

### Request Validation

Request validation is standardized and strictly separated from DTOs:

- All validation uses **ozzo-validation** and lives in `internal/validations/<entity>_validation.go` as standalone functions — **never** inside `internal/dtos/`.
- DTOs hold only struct definitions, tags, and an optional `SetDefaults()` for pagination/query fallbacks (no `Validate()` method).
- Each request type gets a function `Validate<Name>(req *dtos.<Name>) error`. The `validations` package imports `dtos`, not the other way around (avoids an import cycle).
- Controllers wire it explicitly, keeping handlers thin:

```go
func (c *Controllers) ListRoles(ctx *gin.Context) {
    var query dtos.RoleListQuery
    if err := ctx.ShouldBindQuery(&query); err != nil {
        c.wrapError(ctx, constants.ErrBadRequest.WithMessage("invalid query parameters"))
        return
    }
    query.SetDefaults()
    if err := validations.ValidateRoleListQuery(&query); err != nil {
        c.wrapError(ctx, constants.ErrBadRequest.WithMessage(err.Error()))
        return
    }
    resp, err := c.svc.ListRoles(ctx.Request.Context(), query)
    if err != nil {
        c.wrapError(ctx, err)
        return
    }
    c.respondOK(ctx, "Roles retrieved successfully", resp)
}
```

```go
// internal/validations/role_validation.go
func ValidateRoleListQuery(q *dtos.RoleListQuery) error {
    return validation.ValidateStruct(q,
        validation.Field(&q.Page, validation.Min(1)),
        validation.Field(&q.Limit, validation.Min(1), validation.Max(100)),
        validation.Field(&q.Search, validation.Length(0, 100)),
    )
}
```

### DTO Naming & Layering

DTO structs follow two roles:

- Input: `<Name>Request` (JSON body) or `<Name>Query` (query-string/list params; usually has `SetDefaults()`).
- Output: `<Name>Data` for data returned to the client in the response envelope's `data` field (e.g. `APIResponse[HealthData]`). Nested sub-structs used only inside a `*Data` payload need no suffix.

Which layer may touch DTOs:

- `controllers` bind/validate `*Request`/`*Query` and return `*Data`.
- `services` take `*Request`/`*Query` as input and build `*Data` from domain models.
- `repositories` **never** import `internal/dtos` — they work with `internal/models` entities plus primitives or local filter types, so transport shapes never leak into persistence.

### Templates & Large Text Content

HTML, email bodies and long prompts are never string literals inside `.go` files. They live as embedded template files under `internal/templates/<category>/` with a consistent `.tmpl` extension (`*.html.tmpl` / `*.txt.tmpl`), loaded via `//go:embed` and rendered through Go's template engine. Use `html/template` for HTML output (auto-escaping/XSS-safe) and `text/template` for plain text or prompts. The `.go` files hold only the loader, data structs and render functions.

---

## Tech Stack

| Component | Technology | Description |
|-----------|------------|-------------|
| **Language** | Go 1.25+ | Modern, idiomatic Go |
| **HTTP Framework** | Gin | High-performance HTTP web framework |
| **ORM / Database** | GORM / PostgreSQL | Relational persistence with connection pooling |
| **Caching & KV** | Redis | In-memory key-value caching with JSON serialization |
| **Message Broker** | RabbitMQ / amqp091-go | AMQP message broker with topic/fanout exchanges & DLQ |
| **Object Storage** | MinIO / AWS S3 | S3-compatible file storage adapter |
| **Email Service** | net/smtp & Mailpit | Transactional email delivery with local web UI |
| **Migration Tool** | golang-migrate | Clean, versioned database schema migrations |
| **Authentication** | golang-jwt/jwt/v5 | Secure JWT access & refresh token management |
| **Validation** | ozzo-validation | Type-safe declarative struct validation |
| **API Documentation** | Scalar / OpenAPI 3.0 | Modern interactive API documentation |

---

## Prerequisites

- **Go**: 1.25 or higher
- **Docker & Docker Compose**: (Recommended for running PostgreSQL, Redis, RabbitMQ, MinIO, and Mailpit)
- **Make**: (Optional, for running convenient short commands)

---

## Environment Setup

1. Copy `.env.example` to create your local `.env`:
   ```bash
   cp .env.example .env
   ```

2. Configure environment settings:
   ```dotenv
   # Application
   APP_NAME=code-base-golang
   APP_ENV=development
   HTTP_PORT=8080

   # Database (PostgreSQL)
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=postgres
   POSTGRES_DB=code_base_golang
   POSTGRES_SSLMODE=disable

   # Database Auto-Migration on startup
   AUTO_MIGRATE=true

   # Redis Caching
   REDIS_HOST=localhost
   REDIS_PORT=6379

   # RabbitMQ Event Streaming
   RABBITMQ_HOST=localhost
   RABBITMQ_PORT=5672
   RABBITMQ_USER=guest
   RABBITMQ_PASSWORD=guest
   RABBITMQ_VHOST=/

   # S3 / MinIO Object Storage
   S3_ENDPOINT=localhost:9000
   S3_ACCESS_KEY=minioadmin
   S3_SECRET_KEY=minioadmin
   S3_BUCKET_NAME=code-base-golang

   # SMTP Email Delivery (Mailpit local default)
   SMTP_HOST=localhost
   SMTP_PORT=1025
   SMTP_FROM_EMAIL=no-reply@example.com
   SMTP_FROM_NAME="Codebase API"

   # Security & Middlewares
   RATE_LIMIT_ENABLED=true
   RATE_LIMIT_RPS=20
   RATE_LIMIT_BURST=40
   CORS_ALLOWED_ORIGINS=*
   ```

---

## Running the Application

### 1. Start Infrastructure via Docker Compose
Starts PostgreSQL, Redis, RabbitMQ, MinIO, and Mailpit in the background:
```bash
make docker-up
# or: docker compose -f deployment/docker-compose.yaml up -d
```

### 2. Run the API Locally
```bash
make run
# or: go run ./cmd/api
```

---

## Database Migrations & Auto-Migration

Database migrations are managed through a **single unified binary** (`cmd/api/main.go migrate <cmd>`), eliminating the need for external CLI tool installations.

### Common Migration Commands

| Task | Make Command | Direct CLI Command |
|------|--------------|--------------------|
| **Create Migration** | `make migrate-create name=create_users_table` | `go run cmd/api/main.go migrate create create_users_table` |
| **Apply All Migrations** | `make migrate-up` | `go run cmd/api/main.go migrate up` |
| **Apply N Migrations** | `make migrate-up step=1` | `go run cmd/api/main.go migrate up 1` |
| **Rollback 1 Migration** | `make migrate-down` | `go run cmd/api/main.go migrate down 1` |
| **Rollback N Migrations** | `make migrate-down step=2` | `go run cmd/api/main.go migrate down 2` |
| **Rollback All** | `make migrate-down step=all` | `go run cmd/api/main.go migrate down all` |
| **Check Version / Status** | `make migrate-status` | `go run cmd/api/main.go migrate status` |
| **Force Version (Dirty Fix)**| `make migrate-force version=20260928120000` | `go run cmd/api/main.go migrate force 20260928120000` |

### Auto-Migration on Startup

When `AUTO_MIGRATE=true` is set in `.env`:
1. The application automatically detects pending `.up.sql` files in `migrations/`.
2. Migrations execute cleanly upon boot right after database connection is established.
3. If no new migrations exist, it safely skips (`ErrNoChange`) and starts the HTTP server immediately.

---

## Interactive API Documentation

- **Scalar Interactive UI**: [`http://localhost:8080/docs`](http://localhost:8080/docs)
- **OpenAPI 3.0 Specification**: [`http://localhost:8080/openapi.yaml`](http://localhost:8080/openapi.yaml)

> [!NOTE]
> The documentation endpoints are secured with HTTP Basic Authentication via declarative route middleware (`middlewares: [basic_auth]` in `routes.yaml`). Configure credentials via `BASIC_AUTH_USERNAME` and `BASIC_AUTH_PASSWORD` in your `.env`.

---

## Built-in Infrastructure & Middlewares

| Middleware / Component | Description |
|------------------------|-------------|
| **Structured Logging** | Emits JSON logs (`LogLevelError`, `LogLevelWarn`, `LogLevelInfo`), tracks request latency, and assigns/propagates `X-Request-ID`. |
| **Panic Recovery** | Catches unhandled panics, logs stack traces server-side, and outputs standardized safe HTTP 500 JSON error envelopes (`constants.ErrInternalServerError`). |
| **Declarative Route Auth** | Dispatches `auth` (standard user tokens), `anon_auth` (guest tokens), or `optional_auth` (opt-in context injection) directly from `routes.yaml`. |
| **Permission Guard (RBAC)** | Enforces granular route permissions (`permission: "resource:action"`) via `RequirePermission`, supporting wildcard (`*`) super-admin roles. |
| **Basic Authentication** | Constant-time credential matching (`BasicAuth`) used for protecting API documentation and internal endpoints. |
| **Body Size Limiter** | Restricts maximum allowed request payload size to prevent memory exhaustion and DoS attacks. |
| **CORS Middleware** | Preflight `OPTIONS` (204 No Content), configurable origins, methods, headers, and credentials. |
| **Rate Limiter** | In-memory token bucket rate limiting per client IP with background garbage collection. |
| **Client Metadata** | Extracts client IP and User-Agent into standard `context.Context` (`ctxmeta`). |
| **Decoupled Ports** | Abstract consumer interfaces in `services` (`FileStorage`, `EventPublisher`, `EmailSender`) and `workers` (`EventSubscriber`). |

---

## Makefile Commands

```bash
make help          # Show available commands
make run           # Run the API server locally
make build         # Compile binary into bin/api
make test          # Run all unit tests across the workspace
make docker-up     # Start PostgreSQL, Redis, RabbitMQ, MinIO, and Mailpit
make docker-down   # Stop all background Docker containers
make migrate-up    # Apply pending database migrations
make migrate-down  # Rollback last migration
make migrate-status# Check current migration version
```

---

## Running Tests

Execute the comprehensive test suite across all packages:

```bash
make test
# or: go test -v ./...
```

---

## Core Endpoints

### Health & Diagnostics
- **Method**: `GET`
- **Path**: `/v1/health`
- **Sample Response**:
```json
{
  "status": "success",
  "code": "SUCCESS",
  "message": "operation completed successfully",
  "data": {
    "version": "0.1.0",
    "git_hash": "dev",
    "uptime": "5m12s",
    "services": {
      "database": "connected",
      "redis": "connected",
      "s3": "connected",
      "rabbitmq": "connected",
      "smtp": "connected"
    }
  },
  "timestamp": "2026-09-28T07:00:00Z"
}
```

---

## Deployment

Refer to [deployment/README.md](deployment/README.md) for the complete two-stage Docker build pipeline, image caching strategies, and production container guidelines.

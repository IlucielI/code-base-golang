# AGENTS.md — AI Developer Guide for `code-base-golang`

This document defines the architectural invariants, conventions, and workflows for AI coding assistants working in this repository.

---

## 1. Tech Stack & Project Identity

- **Language**: Go 1.25+
- **HTTP Engine**: Gin (`github.com/gin-gonic/gin`)
- **Database / ORM**: PostgreSQL with GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`)
- **Cache & KV**: Redis (`github.com/redis/go-redis/v9`)
- **Message Broker**: RabbitMQ (`github.com/rabbitmq/amqp091-go`) with topic/fanout routing & DLQ
- **Object Storage**: AWS S3 / MinIO (`github.com/aws/aws-sdk-go-v2`)
- **Email Delivery**: `net/smtp` + local Mailpit testing
- **Schema Migrations**: `github.com/golang-migrate/migrate/v4`
- **Validation**: `github.com/go-ozzo/ozzo-validation/v4`
- **Documentation**: OpenAPI 3.0 + Scalar UI (`/docs`, `/openapi.yaml`) embedded via `//go:embed`
- **Linter & Quality**: `golangci-lint` (`.golangci.yml`) with GitHub Actions & GitLab CI pipelines

---

## 2. Architectural Invariants

### 2.1. Modular Clean Architecture (Domain Encapsulation)
- Business features live strictly within self-contained modules under `internal/modules/<domain>/` (e.g., `internal/modules/health/`, `internal/modules/docs/`, `internal/modules/metrics/`).
- Each module encapsulates its own domain concerns:
  - `<domain>_controller.go`: HTTP transport handling (embeds `controller.Base`).
  - `<domain>_usecase.go`: Pure business logic and domain orchestration.
  - `<domain>_dto.go`: Input payloads (`*Request`, `*Query`) and output payloads (`*Data`).
  - `<domain>_model.go`: Domain entities and persistence schemas (embedding `model.BaseModel`).
  - `<domain>_validation.go`: Input validation logic using `ozzo-validation`.
- **Decoupled Ports & Interfaces**: Modules and usecases **MUST NEVER** depend on concrete adapters (`internal/adapters/*`). They consume idiomatic Go consumer interfaces defined at the point of consumption (e.g. `health.Pinger`).
- **Shared Infrastructure**: Cross-cutting reusable infrastructure lives under `internal/pkg/`:
  - `internal/pkg/controller/`: Shared `controller.Base` with standardized error wrapping (`WrapError`) and response helpers (`RespondOK`, `RespondCreated`, `RespondEmpty`).
  - `internal/pkg/response/`: Standard envelope types (`APIResponse[T]`, `BaseResponse`).
  - `internal/pkg/pagination/`: Reusable pagination queries, metadata, and database filters (`Query`, `Meta`, `Filter`, `Sort`).
  - `internal/pkg/model/`: Shared GORM entity base (`BaseModel`, `JSONMap`).
  - `internal/pkg/apperror/`: Application error definitions and stack tracing.

### 2.2. Declarative Route Configuration & Dynamic Dispatch
- Routes are declared in `internal/routes/routes.yaml`.
- The router (`internal/routes/router.go`) inspects the active modules registered in `internal/modules/modules.go` via `mods.Handlers()`.
- Routes and handler methods (e.g., `handler: HealthCheck`) are dynamically mapped via reflection without manual Go router wiring or forwarding boilerplate.
- **Route-Level Middlewares**: Middlewares defined in `routes.yaml` (e.g. `auth`, `basic_auth`, `anon_auth`, `optional_auth`) map directly to PascalCase methods on `*Middlewares` (`internal/middlewares/`).
- **Granular Permissions**: Declared via `permission: "<resource>:<action>"` in `routes.yaml` and mapped directly to `mws.RequirePermission(r.Permission)`.

### 2.3. Central Modules Registry & Nil-Receiver Policy
- Application modules are aggregated in `*modules.Modules` (`internal/modules/modules.go`) initialized at startup (`cmd/api/main.go`).
- Container structs are guaranteed non-nil by construction at startup.
- **NEVER** write defensive `if m == nil` or `if c == nil` receiver checks on core container methods. Sub-adapter availability (e.g. optional RabbitMQ or Redis connection) should be guarded at the field level, not at the container pointer level.

---

## 3. Git Workflow & Commit Rules (MANDATORY)

1. **NEVER PUSH DIRECTLY TO `main`**:
   - Pushing directly to the `main` branch is **STRICTLY PROHIBITED**.
   - All changes must go through a dedicated branch and Pull Request (PR).
2. **Feature Branching**:
   - Always create a dedicated branch off `main` before making any code or schema changes:
     - `feat/<feature-name>` for new features
     - `fix/<issue-name>` for bug fixes
     - `chore/<task-name>` for configuration, docs, or maintenance
     - `refactor/<target>` for code restructuring
3. **Pre-Commit Verification & Atomic Commits (MANDATORY)**:
   - **MANDATORY PRE-COMMIT CHECKS**: Before committing any changes, the assistant **MUST** run `make check` (or `make lint`, `make test`, and `make build`). Commits **MUST NEVER** be created if static analysis / linting fails or if any unit test regresses.
   - Commits must be small, focused, and scoped to a single logical change.
   - **NEVER** combine multiple unrelated changes in a single commit.
   - Follow Conventional Commits: `feat(...)`, `fix(...)`, `refactor(...)`, `build(...)`, `chore(...)`, `test(...)`, `ci(...)`, `docs(...)`.
4. **Pull Request (PR) Mandate**:
   - Push feature branch to origin: `git push -u origin <branch-name>`.
   - Open a PR to `main` using `gh pr create --base main --head <branch-name> --title "..." --body "..."`.
   - Verify `make check` (`make lint`, `make test`, `make build`) passes with zero regressions and zero linter warnings.
   - **MERGE REQUIRES USER APPROVAL**: Leave the PR **OPEN** after creation. The AI assistant must **NEVER** run `gh pr merge` autonomously. The assistant is allowed to merge only after asking for and receiving explicit approval from the USER.

---

## 4. Core Coding Conventions

0. **Module File Naming Suffixes & Organization**:
   - Within each module `internal/modules/<domain>/`:
     - `<domain>_controller.go`: HTTP handler methods.
     - `<domain>_usecase.go`: Business logic implementation.
     - `<domain>_dto.go`: Request and response DTO structs.
     - `<domain>_model.go`: Domain entity and GORM database model.
     - `<domain>_validation.go`: ozzo-validation rules.
   - Shared packages in `internal/pkg/`:
     - `controller/`: Shared controller base methods (`base.go`).
     - `response/`: Response envelopes (`response.go`).
     - `pagination/`: Pagination structures (`pagination.go`).
     - `model/`: Base models (`base_model.go`).
0a. **DTO Naming & Layering (MANDATORY)**:
   - Input payloads: `<Name>Request` for JSON bodies, `<Name>Query` for query params (usually carries `SetDefaults()`).
   - Output payloads: `<Name>Data` for structs returned in `response.APIResponse[T].Data`.
   - Controllers bind and validate `<Name>Request` / `<Name>Query` and emit `<Name>Data`.
   - Usecases accept `<Name>Request` / `<Name>Query` and return `<Name>Data`.
0b. **Shared Controller Base (MANDATORY)**:
   - Every domain controller **MUST** embed `controller.Base` (`internal/pkg/controller/base.go`):
     ```go
     type Controller struct {
         controller.Base
         uc *Usecase
     }
     ```
   - Standard response methods provided by `controller.Base`:
     - `c.RespondOK(ctx, "message", data)`
     - `c.RespondCreated(ctx, "message", data)`
     - `c.RespondEmpty(ctx, "message")`
     - `c.WrapError(ctx, err)`
1. **Error Handling by Layer**:
   - Use `pkg/apperror.AppError` for domain and HTTP errors.
   - Format all HTTP responses using `response.APIResponse` and `response.BaseResponse`.
   - Repository/adapter calls return raw errors; usecases map expected errors (e.g. `gorm.ErrRecordNotFound` $\rightarrow$ `constants.ErrNotFound`) and wrap unexpected failures with `constants.ErrInternalServerError.Wrap(err)`.
   - Controllers invoke `c.WrapError(ctx, err)` which automatically records traces on OpenTelemetry spans and formats standardized fail/error responses.
2. **Context & Metadata Propagation**:
   - Always accept `context.Context` as the first parameter in repository, usecase, and adapter methods.
   - Extract client metadata using `pkg/ctxmeta` (`RequestID`, `ClientIP`, `UserAgent`).
3. **Request Validation (ozzo-validation)**:
   - Request validations live in `<domain>_validation.go` inside the domain module.
   - Expose validation as a standalone function: `func Validate<Name>(req *<Name>Request) error`.

---

## 5. Development Workflows for Agents

### Adding a New Modular Feature
1. **Migration**: Run `make migrate-create name=create_<domain>_table` and write `.up.sql` & `.down.sql`.
2. **Module Directory**: Create `internal/modules/<domain>/`.
3. **Model**: Define entity struct embedding `model.BaseModel` in `<domain>_model.go`.
4. **DTO**: Define payload structs in `<domain>_dto.go`.
5. **Validation**: Define validation functions in `<domain>_validation.go`.
6. **Usecase**: Define business logic in `<domain>_usecase.go`.
7. **Controller**: Define handler struct embedding `controller.Base` in `<domain>_controller.go`.
8. **Module Registry**: Register the new controller in `internal/modules/modules.go` and include in `Handlers() []any`.
9. **Routes**: Register path, HTTP method, and middleware in `internal/routes/routes.yaml`.
10. **Unit Tests**: Add tests alongside code (`<domain>_controller_test.go`, `<domain>_usecase_test.go`).

---

## 6. Development & Quality Assurance Commands

Always run quality checks to verify zero regressions and clean static analysis before committing and concluding tasks:

```bash
# Complete pre-commit check (lint + unit tests + binary build)
make check

# Run static analysis linter (golangci-lint)
make lint

# Run all unit tests
make test
# or: go test -v ./...

# Build binary
make build

# Build three-tier container images
make build-docker

# Start infrastructure services (PostgreSQL, Redis, RabbitMQ, MinIO, Mailpit)
make docker-up

# Stop all background Docker services
make docker-down

# Database migrations
make migrate-create name=<migration_name>
make migrate-up
make migrate-down
make migrate-status
```

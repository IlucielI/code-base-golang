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

### 2.1. Clean Architecture & Decoupled Ports
- Pure business logic lives strictly in `internal/services/`.
- Services and workers **MUST NEVER** import or depend directly on concrete adapters (`internal/adapters/*`).
- Always consume idiomatic Go consumer interfaces defined in domain packages:
  - `services.FileStorage` (`internal/services/storage.go`)
  - `services.EventPublisher` (`internal/services/publisher.go`)
  - `services.EmailSender` (`internal/services/mailer.go`)
  - `workers.EventSubscriber` (`internal/workers/subscriber.go`)
- Concrete adapters (`database`, `redis`, `s3`, `rabbitmq`, `smtp`) live in `internal/adapters/` and are injected in `cmd/api/main.go`.

### 2.2. Declarative Route Configuration & Middlewares
- Routes are declared in `internal/routes/routes.yaml`.
- The router maps handlers, HTTP methods, middlewares, and permissions from this YAML file without manual Go router wiring.
- Register any new endpoints in `routes.yaml` rather than manually calling `router.GET(...)` in Go code.
- **Route-Level Middlewares**: Middlewares defined in `routes.yaml` (e.g. `auth`, `basic_auth`, `anon_auth`, `optional_auth`) must have the direct signature `func(*gin.Context)` on `*Middlewares` (`internal/middlewares/`). The router converts YAML snake_case names to PascalCase methods via `strutil.ToPascalCase` (`internal/pkg/strutil`) and binds them without reflection call overhead.
- **Granular Permissions**: Declared via `permission: "<resource>:<action>"` in `routes.yaml` and mapped directly to `mws.RequirePermission(r.Permission)`.
- **API Documentation Protection**: The OpenAPI specification (`/openapi.yaml`) and interactive Scalar UI (`/docs`) are protected via `middlewares: [basic_auth]` declared directly in `routes.yaml`.

### 2.3. Database Migrations
- Migration files live in `migrations/` named `<timestamp>_<name>.up.sql` and `<timestamp>_<name>.down.sql`.
- When `AUTO_MIGRATE=true`, migrations apply automatically on application/container startup after database connectivity is confirmed.

### 2.4. Container Construction & Nil-Receiver Policy
- Application container structs (`*Service`, `*Repositories`, `*Controllers`, `*Middlewares`, `*WorkerServer`) are guaranteed non-nil by construction at startup (`cmd/api/main.go`).
- **NEVER** write defensive `if s == nil`, `if r == nil`, or `if w == nil` receiver checks on container methods or getters. If a container is nil, it represents a fatal initialization bug that must fail fast, not silently return nil.
- Sub-adapter availability (e.g. optional RabbitMQ broker or Redis connection) should be guarded at the field level (e.g. `if r.rdb == nil` or `if w.sub == nil`), not at the container pointer level.

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

0. **File Naming Suffixes & Constants Conventions (MANDATORY)**:
   - Files in the following packages **MUST** carry a role suffix so each file's layer is unambiguous. This applies to every file in the package, including shared/base ones (e.g. the response envelope and base model), for consistency.
     - `internal/dtos/` → `*_dto.go` (e.g. `recording_dto.go`, `response_dto.go`, `pagination_dto.go`)
     - `internal/models/` → `*_model.go` (e.g. `recording_model.go`, `base_model.go`)
     - `internal/validations/` → `*_validation.go` (e.g. `recording_validation.go`)
   - Test files append the suffix before `_test.go` (e.g. `recording_dto_test.go`, `base_model_test.go`).
   - Rationale: these layers hold many small per-entity files; the suffix keeps the editor's file list self-describing and prevents name collisions across layers (e.g. a `recording` DTO vs. a `recording` model).
   - **No suffix elsewhere.** Other packages name files by topic, not role:
     - `internal/constants/` uses one file per topic/domain (`error.go`, `response.go`, `worker.go`, `auth.go`, `logger.go`, `<domain>.go`); do **not** create `*_constant.go` (it stutters with the package name).
     - Magic strings/numbers for token types (`constants.TokenType*`), user roles (`constants.Role*`), and log levels (`constants.LogLevel*`) **MUST** be defined in `internal/constants/` rather than hardcoded in services, middlewares, or handlers.
0a. **DTO Naming & Layering (MANDATORY)**:
   - **Struct naming** — two roles only:
     - Input payloads: `<Name>Request` for JSON bodies, `<Name>Query` for query-string/list params (the latter usually carries `SetDefaults()`).
     - Output payloads: `<Name>Data` for any struct returned to the client in the response envelope's `data` field (matches `APIResponse[T].Data`, e.g. `APIResponse[HealthData]`). Nested sub-structs used only inside a `*Data` payload need no suffix (e.g. a `HealthServices` embedded in `HealthData`).
   - **Layering — who may use DTOs:**
     - `controllers` consume `*Request`/`*Query` (bind + validate) and emit `*Data`.
     - `services` accept `*Request`/`*Query` as use-case input and build `*Data` from domain models.
     - `repositories` **MUST NEVER** import `internal/dtos`. They operate on `internal/models` entities plus primitives or local filter types (e.g. a `RoleFilter` defined in `repositories`). Transport shapes never leak into the persistence layer.
0b. **Repository Filter & Pagination Conventions (MANDATORY)**:
   - Query filters used by repositories live directly inside `internal/repositories/` — **NEVER** in a sub-package (e.g. no `repositories/filters/`).
   - Repository filters are **pure data structs without logic or calculation methods**. Transport-to-persistence math (such as calculating offset, default limits, or metadata) belongs strictly in `internal/dtos/` and `internal/services/`.
   - Common pagination and sorting parameters use shared types in `internal/repositories/filter.go`:
     - `repositories.PaginationFilter` (`Limit int`, `Offset int`)
     - `repositories.SortFilter` (`SortBy string`, `Order string`)
   - Reusable pagination query and metadata live in `internal/dtos/pagination_dto.go`:
     - `dtos.PaginationQuery` (with `SetDefaults()` and `Offset()`)
     - `dtos.PaginationMeta` (with `CalculateTotalPages()` and `NewPaginationMeta(page, limit, total)`)
   - Domain-specific filter structs live in their respective repository files (e.g. `UserFilter` in `internal/repositories/user_repository.go`) embedding `PaginationFilter` and/or `SortFilter`.
   - Services convert transport queries (`dtos.<Entity>ListQuery`) into `repositories.<Entity>Filter` before invoking the repository:
     ```go
     filter := repositories.UserFilter{
         PaginationFilter: repositories.PaginationFilter{
             Limit:  query.Limit,
             Offset: query.Offset(),
         },
         Search: query.Search,
     }
     ```
1. **Standard Error Envelope & Error Handling by Layer**:
   - Use `pkg/apperror.AppError` for domain and HTTP errors.
   - Format all HTTP responses using the standardized envelope in `internal/dtos/response_dto.go` (`status`, `code`, `message`, `data`, `errors`, `timestamp`).
   - **Repositories return raw errors.** A repository returns the underlying driver/GORM error unchanged (e.g. `return nil, err`, propagating `gorm.ErrRecordNotFound`). Repositories **MUST NOT** import `constants`/`apperror` or assign HTTP status — persistence does not know about transport concerns.
   - **Services translate errors into `AppError`.** The service layer is where raw errors become domain errors. Map expected conditions to the right template (e.g. `errors.Is(err, gorm.ErrRecordNotFound)` → `constants.ErrNotFound`, a 4xx that carries no stack) and wrap everything unexpected as a 5xx: `return constants.ErrInternalServerError.Wrap(err)`.
   - **Why wrap in the service, not the controller.** `AppError.Wrap` captures a Go stack trace for 5xx errors at the *first* wrap point. Wrapping in the service makes the trace point at the service line closest to the failing repository call. If a raw error bubbles all the way up, the controller's `wrapError` wraps it there instead and the stack points at the controller — much less useful. Go cannot reconstruct the original repository line from a bare error, so the service call site is the most useful anchor; do not add `AppError` wrapping inside repositories to chase it.
   - The controller's `wrapError` is the final safety net: it records the error (and any captured stack) on the active OpenTelemetry span, and in development logs the stack for local debugging.
   - Example (service):
     ```go
     role, err := s.repo.FindRoleByID(ctx, id)
     if err != nil {
         if errors.Is(err, gorm.ErrRecordNotFound) {
             return nil, constants.ErrNotFound.WithMessage("role not found")
         }
         return nil, constants.ErrInternalServerError.Wrap(err) // stack captured here
     }
     ```
2. **Context & Metadata Propagation**:
   - Always accept `context.Context` as the first parameter in repository, service, and external adapter methods.
   - Extract client metadata using `pkg/ctxmeta` (`RequestID`, `ClientIP`, `UserAgent`).
3. **Request Validation (MANDATORY — ozzo-validation, in `internal/validations/`)**:
   - All request validation **MUST** use `ozzo-validation` (`github.com/go-ozzo/ozzo-validation/v4`) and **MUST** live in `internal/validations/<entity>_validation.go` — **never** inside `internal/dtos/`.
   - DTOs (`internal/dtos/*_dto.go`) contain **only** struct definitions, JSON/form/query tags, and an optional `SetDefaults()` method for pagination/query fallbacks. They **MUST NOT** carry a `Validate()` method or any validation rules.
   - Expose validation as a standalone function per request type: `func Validate<Name>(req *dtos.<Name>) error` that returns `validation.ValidateStruct(req, validation.Field(...), ...)`. The `validations` package imports `dtos`; `dtos` **MUST NOT** import `validations` (avoids an import cycle).
   - Controllers invoke validation explicitly; business logic never lives in the controller. The canonical handler flow is:
     ```go
     func (c *Controllers) ListRoles(ctx *gin.Context) {
         var query dtos.RoleListQuery
         if err := ctx.ShouldBindQuery(&query); err != nil { // or ShouldBindJSON for bodies
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
   - Example validation definition in `internal/validations/role_validation.go`:
     ```go
     func ValidateRoleListQuery(q *dtos.RoleListQuery) error {
         return validation.ValidateStruct(q,
             validation.Field(&q.Page, validation.Min(1)),
             validation.Field(&q.Limit, validation.Min(1), validation.Max(100)),
             validation.Field(&q.Search, validation.Length(0, 100)),
         )
     }
     ```
4. **Resource Safety**:
   - Always ensure connections, response bodies, and lock releases are deferred cleanly (`defer`).
   - Workers and async subscribers must handle graceful cancellation via `ctx.Done()`.
5. **Templates & Large Text Content (MANDATORY)**:
   - HTML, email bodies, long prompts, or any multi-line markup/text block **MUST NOT** be embedded as string literals inside `.go` files. Store them as template files under `internal/templates/<category>/` (e.g. `emails/`, `prompts/`).
   - Use a **consistent `.tmpl` extension** so the embed glob never silently skips a file: `*.html.tmpl` for HTML, `*.txt.tmpl` for plain text/prompts.
   - Load templates with `//go:embed` and render them through Go's template engine. The `.go` file holds only the embed directive, the parsed-template variables, the data structs, and thin render functions — never the content itself.
   - Choose the engine by output type for safety: **`html/template` for anything rendered as HTML** (it auto-escapes and prevents injection/XSS), and `text/template` only for non-HTML output (plain-text emails, LLM prompts).
   - Example:
     ```go
     //go:embed emails/*.tmpl
     var emailFS embed.FS

     var emailHTML = htmltemplate.Must(htmltemplate.ParseFS(emailFS, "emails/*.html.tmpl"))

     func RenderWelcomeEmail(data WelcomeEmailData) (string, error) {
         var buf bytes.Buffer
         if err := emailHTML.ExecuteTemplate(&buf, "welcome.html.tmpl", data); err != nil {
             return "", fmt.Errorf("execute welcome.html.tmpl: %w", err)
         }
         return strings.TrimSpace(buf.String()), nil
     }
     ```

---

## 5. Development Workflows for Agents

### Adding a New Entity & CRUD Feature
1. **Migration**: Run `make migrate-create name=create_<entity>_table` and write `.up.sql` & `.down.sql`.
2. **Model**: Create entity struct embedding `models.BaseModel` in `internal/models/<entity>_model.go`.
3. **Repository**: Define interface and GORM queries in `internal/repositories/`.
4. **DTO**: Define request/response payload structs in `internal/dtos/<entity>_dto.go` (struct + tags only; add `SetDefaults()` for list/query DTOs). No `Validate()` method here.
5. **Validation**: Define `func Validate<Name>(req *dtos.<Name>) error` using `ozzo-validation` in `internal/validations/<entity>_validation.go`.
6. **Service**: Define interface & business implementation in `internal/services/`.
7. **Controller**: Add handler method in `internal/controllers/` following the flow: bind (`ShouldBindJSON`/`ShouldBindQuery`) → `SetDefaults()` → `validations.Validate<Name>(&req)` → service call → `respondOK`/`respondCreated`.
8. **Routes**: Register path, HTTP method, and auth middlewares in `internal/routes/routes.yaml`.
9. **Unit Tests**: Add tests alongside code using Go's standard `testing` package.

### Adding an Asynchronous Event / Worker
1. Define the queue message payload struct in `internal/queue/`.
2. Publish the event via `services.EventPublisher.Publish(ctx, exchange, routingKey, body)`.
3. Add the handler method in `internal/workers/` and subscribe it inside `WorkerServer.RegisterWorker` with `w.sub.MustSubscribe(constants.Topic<Name>, w.Handle<Name>)`.

**Worker handler rules (keep handlers thin, no boilerplate):**
- **Never** repeat the parse → log → run-service → map-to-ack sequence in each handler. Route every handler through the shared generic `handleMessage[T]` helper in `worker.go`, which decodes the payload, logs the start/finish/error lifecycle uniformly, and returns the ack/nack bool:
  ```go
  func (w *WorkerServer) HandleUserRegistered(ctx context.Context, msg []byte) bool {
      return handleMessage(ctx, "user.registered", msg, w.svc.NotifyUserRegistered)
  }
  ```
  The `process` function passed in is typically a service method with signature `func(context.Context, T) error`; business logic stays in the service, never in the handler.
- A handler returns `true` to ACK and `false` to NACK (routing the message to the DLQ/retry). Let `handleMessage` derive this from the service error — do not hand-roll ack logic.
- Always honour the broker-provided `ctx` (for cancellation/tracing); do not create a fresh `context.Background()` inside handlers.
- **One handler file per domain**, not one growing `worker.go`. `worker.go` holds only `WorkerServer`, `New`, `RegisterWorker` and the shared helper; put domain handlers in `internal/workers/<domain>_worker.go` (e.g. `recording_worker.go`). `RegisterWorker` stays the single registry of all subscriptions, grouped by domain with comments.

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

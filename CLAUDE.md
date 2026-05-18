# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PokeForum is a RESTful API backend for a modern forum/community platform, built with Go 1.25.5. It follows a layered clean architecture with clear separation between controllers, services, and repositories.

**Tech Stack:**
- **Web Framework:** Gin (v1.11.0)
- **ORM:** Entgo (v0.14.5) - Facebook's entity framework for Go
- **Database:** PostgreSQL with `pg_trgm` extension for full-text search
- **Cache/Queue:** Redis (v9.17.2)
- **Auth:** Sa-Token (v0.1.7)
- **DI Container:** samber/do (v1.6.0)
- **Async Tasks:** hibiken/asynq (v0.25.1)

## Architecture

The codebase follows a strict **layered architecture** with dependency injection:

```
HTTP Request → Middleware → Controller → Service → Repository → Ent ORM → PostgreSQL
```

### Key Layers

1. **Controller Layer** (`/internal/controller/`): HTTP handlers with Swagger annotations
2. **Service Layer** (`/internal/service/`): Business logic with interface-based design
3. **Repository Layer** (`/internal/repository/`): Data access abstraction over Ent ORM
4. **Middleware** (`/internal/middleware/`): Gin middleware (logging, recovery, CORS, rate limiting, auth)
5. **Dependency Injection** (`/internal/initializer/do_service.go`): Constructor-based DI using samber/do

### Database Schema

Database schemas are defined in `/ent/schema/` (22 entities). **Key design decisions:**
- No foreign keys at database level (app-layer referential integrity)
- GIN indexes on `posts.title` for fuzzy full-text search
- All entities have timestamps (TimeMixin)
- Composite indexes for query optimization

**Core Entities:**
- User management: `users`, `user_login_logs`, `user_balance_logs`, `user_signin_status`, `user_oauths`
- Forum content: `categories`, `posts`, `post_actions`, `comments`, `comment_actions`
- Moderation: `category_moderators`, `blacklists`, `settings`, `oauth_providers`

## Essential Commands

### Development

```bash
# Run development server with debug mode and Swagger UI
make run

# Generate Ent ORM code after schema changes
make gen

# Generate Swagger documentation
make docs
```

### Code Quality

```bash
# Run linter (golangci-lint)
make lint

# Auto-fix linting issues
make lint-fix

# Format code with gofmt and goimports
make fmt

# Run go vet
make vet
```

### Testing

```bash
# Run all tests with race detector
make test

# Run a specific test
# Example: go test -v ./internal/service/... -run TestUserService

# Generate coverage report (coverage.html)
make test-coverage

# Run short tests only (skip long-running tests)
make test-short
```

### Building

```bash
# Build Linux AMD64 with UPX compression
make build

# Build Linux ARM64
make build-linux-arm64

# Build all platforms
make build-all

# Clean build artifacts
make clean
```

### Dependencies

```bash
# Download dependencies
make deps

# Tidy go.mod
make deps-tidy

# Update all dependencies
make deps-update

# Install dev tools (golangci-lint, swag, goimports)
make install-tools
```

### CI/CD

```bash
# Run full CI pipeline
make ci

# Pre-commit checks (fmt + lint + test-short)
make pre-commit
```

## Setup

### Database Setup

Enable the GIN index extension for full-text search:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

### Configuration

Copy `config.example.yaml` to `config.yaml` and configure:
- PostgreSQL connection (host, port, database, user, password)
- Redis connection (host, port, db)

## Adding New Features

The standard workflow for adding a new feature (e.g., a new entity):

1. **Define Schema** (`/ent/schema/entity.go`):
   ```bash
   # Edit schema, then generate Ent code
   make gen
   ```

2. **Create Repository** (`/internal/repository/entity_repository.go`):
   - Implement CRUD operations using Ent client
   - Return interfaces for testability

3. **Create Service** (`/internal/service/entity_service.go`):
   - Define interface `IEntityService`
   - Implement business logic
   - May use cache, async tasks, or other services

4. **Create Controller** (`/internal/controller/entity_controller.go`):
   - Define request/response DTOs in `/internal/schema/`
   - Add Swagger annotations
   - Validate requests and call service

5. **Register DI** (`/internal/initializer/do_service.go`):
   ```go
   do.Provide(injector, func(i *do.Injector) (service.IEntityService, error) {
       repos := do.MustInvoke[*repository.Repositories](i)
       logger := do.MustInvoke[*zap.Logger](i)
       cacheService := do.MustInvoke[cache.ICacheService](i)
       return service.NewEntityService(repos.Entity, cacheService, logger), nil
   })
   ```

6. **Add Routes** (`/internal/initializer/router.go`):
   ```go
   entityService := do.MustInvoke[service.IEntityService](injector)
   entityCon := controller.NewEntityController(entityService)
   entityCon.EntityRouter(api.Group("/entities"))
   ```

7. **Update Documentation**:
   ```bash
   make docs
   ```

## API Structure

Base URL: `/api/v1`

### Route Groups

- **Public:** `/health`, `/auth/*`, `/config/*`, `/oauth/*`
- **Authenticated:** `/profile/*`, `/ranking/*`, `/categories/*`, `/posts/*`, `/comments/*`, `/signin/*`
- **Moderator (role-protected):** `/moderator/*`
- **Admin (role-protected):** `/manage/dashboard/*`, `/manage/users/*`, `/manage/categories/*`, `/manage/posts/*`, `/manage/comments/*`
- **Super Admin:** `/super/manage/settings/*`, `/super/manage/settings/oauth/*`

### Authentication

Uses Sa-Token for session management with 4 user roles:
- User (default)
- Moderator
- Admin
- SuperAdmin

Routes are protected using `saLoginCheck` middleware in `router.go`.

## Special Features

### Rate Limiting

Custom Redis-based sliding window algorithm with multiple tiers:
- Global: 100 req/sec
- API: 60 req/min
- Auth: 10 req/min (brute force protection)

Degraded behavior: allows requests if Redis fails.

### Async Tasks

Uses Asynq for background job processing (Redis-backed):
- Sign-in reward distribution
- Statistics synchronization (every 5 minutes)
- Task handlers in `/internal/pkg/asynq/`

### Email System

SMTP-based email sending with connection pooling:
- Password reset emails
- HTML email templates
- Configured via settings

### Statistics

Background stats sync runs every 5 minutes via Asynq. Separate stats services for posts/comments.

## Code Quality Standards

### Linting Configuration

Uses golangci-lint with comprehensive ruleset (see `.golangci.yml`):
- Standard linters enabled
- Additional linters: bodyclose, dogsled, dupl, goconst, gocritic, gosec, misspell, nakedret, noctx, prealloc, unconvert, unparam, whitespace
- Formatters: gofmt, goimports
- Exclusions: Generated code (ent/, docs/), test files

### Code Style

- Bilingual comments (English + Chinese)
- Interface-first design for services and repositories
- Structured logging with uber-go/zap
- Clear separation of concerns

## Important Notes

- **Debug mode:** Enable with `--debug` flag (enables Swagger UI at `/swagger/*`, Gin debug mode)
- **Migrations:** Auto-run on startup via Ent
- **Foreign keys:** None at database level - enforced in application layer
- **Testing:** Use `-short` flag to skip long-running tests
- **Build compression:** UPX used for production binaries (can fail silently, handled with `|| true`)
- **Prometheus metrics:** Optional (configurable via config)

## Entry Points

- **Server startup:** `/cmd/server.go` → `RunServer()`
- **Router setup:** `/internal/initializer/router.go`
- **DI container:** `/internal/initializer/do_service.go`
- **Config loading:** `/internal/initializer/viper.go`
- **Database init:** `/internal/initializer/db.go`

## Task Guidelines

- Project dependencies: Go + Gin + EntORM
- Project middleware dependencies: PostgreSQL + Redis
- Parameter validation implemented using github.com/go-playground/validator/v10 library
- Request/response body code must be created under the Schema directory
- Interfaces require comments compliant with gin-swagger specifications for API documentation generation
- API interfaces must adhere to REST API style
- Strictly prohibit foreign key associations in database table design; all related logic must be implemented via logical queries at the application layer
- Code comment structure rule: {English} | {Chinese}
- The tracing.WithTraceIDField method must be placed as the second argument in s.logger.\*() methods. Placement elsewhere or as the last argument is not permitted. [To standardize log formats]
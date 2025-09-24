# LarGo — Laravel-style Go Web Framework

Early work-in-progress. Goals:

- Cobra-based CLI (`largo`) with scaffolding and generators
- HTTP runtime with router, middleware, context, responses
- Config/env loader, logging, validation
- Migrations and DB helpers
- Packaging via GoReleaser + Homebrew/Scoop + Docker

## Building the CLI

go build -ldflags "-X main.version=v0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/largo

## Usage

- `largo --help`
- `largo version`
- `largo new <app>`
- `largo serve [target]` (dev runner for generated app)
- `largo make:controller <Name>`
- `largo make:migration <name>`
- `largo make:model <Name>`
- `largo make:middleware <Name>`
- `largo migrate` / `largo migrate:rollback` / `largo migrate:status`

### HTTP runtime

Generated apps use `pkg/httpx` for a minimal, idiomatic router, context, and middleware:

- Exact and param routing (via chi) and middleware chain
- Built-ins: `RequestID`, `Recover`, `Logger`
- Helpers: `Context.JSON`, `Context.Text`, `httpx.ServeEnv` with safe timeouts
  - URL params: `c.Param("id")` from routes like `/users/{id}`

### Config & Env

- `pkg/config` provides `config.Load()` with defaults and `.env`/env precedence
- Keys: `LARGO_ENV`, `PORT`, `DATABASE_URL`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`, `HTTP_MAX_HEADER_BYTES`

### Binding, Validation, Errors

- Binding: `pkg/httpx/binding` with `BindJSON` and `BindQuery`
- Validation: `go-playground/validator` with common rules (`required,email,min,max,len,oneof,url,uuid,numeric,alphanum,gt,gte,lt,lte`)
- Unified errors: `pkg/httpx/xerr` envelope `{error,message,details,request_id}`; used for 400/404/422/500

### Migrations (Postgres for now)

Set `DATABASE_URL` in your environment or a local `.env` file in the project root. Example:

```
DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable
```

Place `.sql` files under `internal/db/migrations`. Files are applied in lexicographic order. Use `-- up` and `-- down` sections:

```
-- up
CREATE TABLE posts(id serial primary key, title text NOT NULL);

-- down
DROP TABLE posts;
```

Commands:

- `largo migrate` — applies all pending migrations in a new batch
- `largo migrate:rollback` — rolls back the last applied batch
- `largo migrate:status` — shows applied vs pending

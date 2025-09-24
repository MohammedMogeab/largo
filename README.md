<div align="center">

# LarGo

Laravel‑style Go Web Framework — batteries included, DX first.

</div>

**Highlights**
- First‑class CLI for scaffolding, running, generating
- Clean runtime (router, context, middleware, responses)
- Typed config from `.env`/env with safe server defaults
- SQL‑first migrations (Postgres; SQLite next)
- Generators for controllers, models, middleware, migrations

**Install**
- `go install github.com/MohammedMogeab/largo/cmd/largo@latest`
- Or build from source: `go build ./cmd/largo`

**Quickstart**
- `largo new blog && cd blog`
- `cp .env.example .env`
- `go run ./cmd/server`
- Try: `/`, `/hello/alice`, `POST /signup` (returns 201 or 422)

**CLI**
- `largo --help`, `largo version`
- `largo new <app>`, `largo serve [target]`
- `largo make:controller|model|middleware|migration`
- `largo migrate` / `migrate:rollback` / `migrate:status`

**HTTP Runtime**
- Router with exact + param routes (chi), JSON 404/405
- Middlewares: `RequestID`, `Recover`, `Logger`
- Helpers: `Context.JSON`, `Context.Text`, `Context.Param`, `httpx.ServeConfig`

**Config & Env**
- `pkg/config` → `config.Load()` with `.env` → env → defaults
- Keys: `LARGO_ENV`, `PORT`, `DATABASE_URL`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`, `HTTP_MAX_HEADER_BYTES`

**Binding, Validation, Errors**
- Binding: `BindJSON`, `BindQuery` (gorilla/schema)
- Validation: `go-playground/validator` with common rules
- Unified JSON error envelope `{error,message,details?,request_id?}` (400/404/422/500)

**Migrations (Postgres)**
- `DATABASE_URL=postgres://user:pass@localhost:5432/myapp?sslmode=disable`
- Place files in `internal/db/migrations` with `-- up` / `-- down` sections
- Commands: `migrate`, `migrate:rollback`, `migrate:status`

**Build with version info**
- `go build -ldflags "-X main.version=v0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/largo`

— Made with Go. Feedback and contributions welcome.


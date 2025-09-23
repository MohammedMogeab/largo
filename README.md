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
- `largo migrate` / `largo migrate:rollback` / `largo migrate:status`

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

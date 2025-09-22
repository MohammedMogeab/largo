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
- `largo new <app>` (stubbed, implementation next)


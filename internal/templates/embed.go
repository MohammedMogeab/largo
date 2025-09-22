package templates

import "embed"

// FS exposes embedded scaffolding templates used by `largo new` and `largo make:*`.
//
//go:embed app/** stubs/**
var FS embed.FS


package main

import (
    "github.com/MohammedMogeab/largo/internal/cli"
)

// These are set via -ldflags at build time:
// -X main.version=v0.1.0 -X main.commit=abcdef -X main.date=2025-09-22
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    cli.Execute(version, commit, date)
}


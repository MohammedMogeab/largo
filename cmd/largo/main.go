package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/MohammedMogeab/largo/internal/cli"
)
 
var (
	// defaults shown when neither build info nor VERSION file are present
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	v, c, d := deriveVersion(version, commit, date)
	cli.Execute(v, c, d)
}

func deriveVersion(v, c, d string) (string, string, string) {
	// 1) Try Go's build info (works great for `go install module@v0.1.0`)
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			v = bi.Main.Version // e.g., v0.1.0
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if c == "none" && len(s.Value) >= 7 {
					c = s.Value[:7]
				}
			case "vcs.time":
				if d == "unknown" && len(s.Value) >= 10 {
					d = s.Value[:10] // YYYY-MM-DD
				}
			}
		}
	}

	// 2) Fallback: read a VERSION file next to the installed binary
	//    (your installer can drop this file into $GOPATH/bin alongside `largo`)
	if v == "dev" || c == "none" || d == "unknown" {
		if vv, cc, dd, ok := readVersionFile(); ok {
			if v == "dev" && vv != "" {
				v = vv
			}
			if c == "none" && cc != "" {
				c = cc
			}
			if d == "unknown" && dd != "" {
				d = dd
			}
		}
	}
	return v, c, d
}

func readVersionFile() (v, c, d string, ok bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", "", false
	}
	dir := filepath.Dir(exe)
	data, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		return "", "", "", false
	}
	// Accept formats:
	//   "v0.1.0"
	//   "v0.1.0 abc1234"
	//   "v0.1.0 abc1234 2025-09-22"
	parts := strings.Fields(strings.TrimSpace(string(data)))
	switch len(parts) {
	case 1:
		return parts[0], "", "", true
	case 2:
		return parts[0], parts[1], "", true
	default:
		return parts[0], parts[1], parts[2], true
	}
}

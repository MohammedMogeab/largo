package cli

import (
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    "text/template"
    "time"

    "github.com/spf13/cobra"
    "github.com/MohammedMogeab/largo/internal/templates"
)

type newOptions struct {
    Module string
    Force  bool
}

func newNewCmd() *cobra.Command {
    opts := newOptions{}
    cmd := &cobra.Command{
        Use:   "new <app>",
        Short: "Scaffold a new LarGo application",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            appName := sanitizeAppName(args[0])
            if appName == "" {
                return errors.New("app name must contain letters or numbers")
            }
            module := opts.Module
            if module == "" {
                module = appName
            }
            targetDir := appName
            if err := scaffoldApp(cmd, targetDir, appName, module, opts.Force); err != nil {
                return err
            }
            fmt.Fprintf(cmd.OutOrStdout(), "\nCreated %s. Next steps:\n", appName)
            fmt.Fprintf(cmd.OutOrStdout(), "  cd %s\n  go run ./cmd/server\n", appName)
            return nil
        },
    }
    cmd.Flags().StringVarP(&opts.Module, "module", "m", "", "Go module path for the new app (default: app name)")
    cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Overwrite existing files if present")
    return cmd
}

func sanitizeAppName(s string) string {
    s = strings.TrimSpace(s)
    s = strings.Trim(s, ".-_")
    // Replace spaces with dashes and collapse multiple separators
    s = strings.ReplaceAll(s, " ", "-")
    // Very light sanity: keep common filename-safe chars
    out := make([]rune, 0, len(s))
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
            out = append(out, r)
        }
    }
    return strings.Trim(string(out), ".-_")
}

func scaffoldApp(cmd *cobra.Command, targetDir, appName, modulePath string, force bool) error {
    // If target exists and not empty, guard unless --force
    if fi, err := os.Stat(targetDir); err == nil && fi.IsDir() && !force {
        // Check if directory is empty
        f, err := os.Open(targetDir)
        if err == nil {
            defer f.Close()
            names, _ := f.Readdirnames(1)
            if len(names) > 0 {
                return fmt.Errorf("directory %q already exists and is not empty (use --force to overwrite)", targetDir)
            }
        }
    }

    data := map[string]any{
        "AppName":    appName,
        "ModulePath": modulePath,
        "Year":       time.Now().Year(),
    }

    // Walk embedded templates under app/
    return fs.WalkDir(templates.FS, "app", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel("app", path)
        dest := filepath.Join(targetDir, trimTmplSuffix(rel))
        if d.IsDir() {
            if rel == "." {
                return os.MkdirAll(targetDir, 0o755)
            }
            return os.MkdirAll(dest, 0o755)
        }

        // Render text templates; all current files are text-based
        b, err := fs.ReadFile(templates.FS, path)
        if err != nil {
            return err
        }
        // Prepare template with basic funcs
        t, err := template.New(filepath.Base(path)).Funcs(template.FuncMap{
            "ToUpper": strings.ToUpper,
            "ToLower": strings.ToLower,
            "Title":   strings.Title,
        }).Parse(string(b))
        if err != nil {
            return fmt.Errorf("parse template %s: %w", path, err)
        }
        // Guard if exists and not forcing
        if !force {
            if _, err := os.Stat(dest); err == nil {
                return fmt.Errorf("file exists: %s (use --force to overwrite)", dest)
            }
        }
        if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
            return err
        }
        f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
        if err != nil {
            return err
        }
        defer f.Close()
        if err := t.Execute(f, data); err != nil {
            return fmt.Errorf("render %s: %w", path, err)
        }
        return nil
    })
}

func trimTmplSuffix(p string) string {
    if strings.HasSuffix(p, ".tmpl") {
        return strings.TrimSuffix(p, ".tmpl")
    }
    return p
}

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

func newMakeControllerCmd() *cobra.Command {
    var (
        outDir = "internal/handlers"
        force  bool
        pkg    string
    )
    cmd := &cobra.Command{
        Use:   "make:controller <Name>",
        Short: "Generate a controller in internal/handlers",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := strings.TrimSpace(args[0])
            if name == "" {
                return errors.New("controller name is required")
            }
            if pkg == "" {
                pkg = filepath.Base(outDir)
            }
            data := map[string]any{
                "Name": name,
                "Package": pkg,
            }
            destFile := filepath.Join(outDir, toSnake(name)+".go")
            return renderStub(cmd, "stubs/controller.go.tmpl", destFile, data, force)
        },
    }
    cmd.Flags().StringVar(&outDir, "dir", outDir, "Output directory for controller")
    cmd.Flags().StringVar(&pkg, "package", pkg, "Package name (default: dirname)")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite the file if it exists")
    return cmd
}

func newMakeMigrationCmd() *cobra.Command {
    var (
        outDir = "internal/db/migrations"
        force  bool
    )
    cmd := &cobra.Command{
        Use:   "make:migration <name>",
        Short: "Generate a SQL migration file",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := strings.TrimSpace(args[0])
            if name == "" {
                return errors.New("migration name is required")
            }
            ts := time.Now().UTC().Format("20060102150405")
            filename := fmt.Sprintf("%s_%s.sql", ts, toSnake(name))
            destFile := filepath.Join(outDir, filename)
            data := map[string]any{
                "Name": name,
                "Timestamp": ts,
            }
            return renderStub(cmd, "stubs/migration.sql.tmpl", destFile, data, force)
        },
    }
    cmd.Flags().StringVar(&outDir, "dir", outDir, "Output directory for migrations")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite if the file exists")
    return cmd
}

func renderStub(cmd *cobra.Command, stubPath, dest string, data map[string]any, force bool) error {
    // Read template from embedded FS
    b, err := fs.ReadFile(templates.FS, stubPath)
    if err != nil {
        return fmt.Errorf("read stub %s: %w", stubPath, err)
    }
    if !force {
        if _, err := os.Stat(dest); err == nil {
            return fmt.Errorf("file exists: %s (use --force to overwrite)", dest)
        }
    }
    if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
        return err
    }
    t, err := template.New(filepath.Base(stubPath)).Funcs(template.FuncMap{
        "ToUpper": strings.ToUpper,
        "ToLower": strings.ToLower,
        "Title":   strings.Title,
    }).Parse(string(b))
    if err != nil {
        return fmt.Errorf("parse template: %w", err)
    }
    f, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    defer f.Close()
    if err := t.Execute(f, data); err != nil {
        return fmt.Errorf("render: %w", err)
    }
    fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", dest)
    return nil
}

// toSnake converts CamelCase or mixed-case to snake_case
func toSnake(s string) string {
    if s == "" {
        return s
    }
    var out []rune
    var prevLower bool
    for _, r := range s {
        if r >= 'A' && r <= 'Z' {
            if prevLower {
                out = append(out, '_')
            }
            out = append(out, r+('a'-'A'))
            prevLower = false
        } else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
            out = append(out, r)
            prevLower = (r >= 'a' && r <= 'z')
        } else {
            // treat separators as underscore
            if len(out) > 0 && out[len(out)-1] != '_' {
                out = append(out, '_')
            }
            prevLower = false
        }
    }
    // trim trailing underscore
    if len(out) > 0 && out[len(out)-1] == '_' {
        out = out[:len(out)-1]
    }
    return string(out)
}


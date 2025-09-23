package cli

import (
    "bufio"
    "context"
    "database/sql"
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/joho/godotenv"
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/spf13/cobra"
)

type migrateOptions struct {
    Dir         string
    DatabaseURL string
}

func newMigrateCmd() *cobra.Command {
    opts := migrateOptions{Dir: "internal/db/migrations"}
    cmd := &cobra.Command{
        Use:   "migrate",
        Short: "Apply all pending database migrations",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withDB(cmd, opts, func(ctx context.Context, db *sql.DB) error {
                return migrateUp(ctx, cmd, db, opts.Dir)
            })
        },
    }
    addMigrateFlags(cmd, &opts)
    return cmd
}

func newMigrateRollbackCmd() *cobra.Command {
    opts := migrateOptions{Dir: "internal/db/migrations"}
    cmd := &cobra.Command{
        Use:   "migrate:rollback",
        Short: "Rollback the last migration batch",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withDB(cmd, opts, func(ctx context.Context, db *sql.DB) error {
                return migrateRollback(ctx, cmd, db, opts.Dir)
            })
        },
    }
    addMigrateFlags(cmd, &opts)
    return cmd
}

func newMigrateStatusCmd() *cobra.Command {
    opts := migrateOptions{Dir: "internal/db/migrations"}
    cmd := &cobra.Command{
        Use:   "migrate:status",
        Short: "Show migration status",
        RunE: func(cmd *cobra.Command, args []string) error {
            return withDB(cmd, opts, func(ctx context.Context, db *sql.DB) error {
                return migrateStatus(ctx, cmd, db, opts.Dir)
            })
        },
    }
    addMigrateFlags(cmd, &opts)
    return cmd
}

func addMigrateFlags(cmd *cobra.Command, opts *migrateOptions) {
    cmd.Flags().StringVar(&opts.Dir, "dir", opts.Dir, "Migrations directory")
    cmd.Flags().StringVar(&opts.DatabaseURL, "database-url", "", "Database URL (overrides env DATABASE_URL)")
}

func withDB(cmd *cobra.Command, opts migrateOptions, fn func(context.Context, *sql.DB) error) error {
    // Load .env if present (ignore errors)
    _ = godotenv.Load()

    dsn := strings.TrimSpace(opts.DatabaseURL)
    if dsn == "" {
        dsn = os.Getenv("DATABASE_URL")
    }
    if dsn == "" {
        return errors.New("DATABASE_URL is not set; pass --database-url or set in .env")
    }
    if !isPostgres(dsn) {
        return fmt.Errorf("unsupported database URL (only postgres is supported for now): %s", redactDSN(dsn))
    }

    // pgx registers as driver name "pgx"
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return err
    }
    defer db.Close()
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("database connection failed: %w", err)
    }

    if err := ensureMigrationsTable(ctx, db); err != nil {
        return err
    }
    return fn(ctx, db)
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
    _, err := db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            name       TEXT PRIMARY KEY,
            batch      INTEGER NOT NULL,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
        )
    `)
    return err
}

type migrationFile struct {
    Name   string
    Path   string
    UpSQL  string
    DownSQL string
}

func readMigrations(dir string) ([]migrationFile, error) {
    var files []string
    err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if strings.HasSuffix(d.Name(), ".sql") {
            files = append(files, path)
        }
        return nil
    })
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return nil, fmt.Errorf("migrations directory not found: %s", dir)
        }
        return nil, err
    }
    sort.Strings(files)
    out := make([]migrationFile, 0, len(files))
    for _, f := range files {
        mf, err := parseMigrationFile(f)
        if err != nil {
            return nil, err
        }
        out = append(out, mf)
    }
    return out, nil
}

func parseMigrationFile(path string) (migrationFile, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return migrationFile{}, err
    }
    name := filepath.Base(path)
    up, down := splitUpDown(string(b))
    return migrationFile{Name: name, Path: path, UpSQL: up, DownSQL: down}, nil
}

func splitUpDown(s string) (up, down string) {
    // Split by lines beginning with "-- up" / "-- down"
    scanner := bufio.NewScanner(strings.NewReader(s))
    var cur *strings.Builder
    var upBuf, downBuf strings.Builder
    section := ""
    for scanner.Scan() {
        line := scanner.Text()
        ltrim := strings.TrimSpace(strings.ToLower(line))
        switch ltrim {
        case "-- up":
            section = "up"
            cur = &upBuf
            continue
        case "-- down":
            section = "down"
            cur = &downBuf
            continue
        }
        if cur == nil {
            // content before any marker goes to up
            cur = &upBuf
            section = "up"
        }
        if section != "" {
            cur.WriteString(line)
            cur.WriteByte('\n')
        }
    }
    return strings.TrimSpace(upBuf.String()), strings.TrimSpace(downBuf.String())
}

func migrateUp(ctx context.Context, cmd *cobra.Command, db *sql.DB, dir string) error {
    files, err := readMigrations(dir)
    if err != nil {
        return err
    }
    applied, err := getApplied(ctx, db)
    if err != nil {
        return err
    }
    pending := make([]migrationFile, 0)
    for _, f := range files {
        if !applied[f.Name] {
            pending = append(pending, f)
        }
    }
    if len(pending) == 0 {
        fmt.Fprintln(cmd.OutOrStdout(), "No pending migrations.")
        return nil
    }
    batch, err := nextBatch(ctx, db)
    if err != nil {
        return err
    }
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    for _, m := range pending {
        if strings.TrimSpace(m.UpSQL) == "" {
            return fmt.Errorf("migration %s has no up SQL", m.Name)
        }
        if _, err := tx.ExecContext(ctx, m.UpSQL); err != nil {
            return fmt.Errorf("apply %s: %w", m.Name, err)
        }
        if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(name, batch) VALUES ($1, $2)`, m.Name, batch); err != nil {
            return err
        }
        fmt.Fprintf(cmd.OutOrStdout(), "Applied %s\n", m.Name)
    }
    if err := tx.Commit(); err != nil {
        return err
    }
    fmt.Fprintf(cmd.OutOrStdout(), "Applied %d migrations in batch %d\n", len(pending), batch)
    return nil
}

func migrateRollback(ctx context.Context, cmd *cobra.Command, db *sql.DB, dir string) error {
    files, err := readMigrations(dir)
    if err != nil {
        return err
    }
    fileMap := make(map[string]migrationFile, len(files))
    for _, f := range files {
        fileMap[f.Name] = f
    }
    last, err := lastBatch(ctx, db)
    if err != nil {
        return err
    }
    if last == 0 {
        fmt.Fprintln(cmd.OutOrStdout(), "Nothing to rollback.")
        return nil
    }
    // Fetch migrations in last batch in reverse lexicographic order
    rows, err := db.QueryContext(ctx, `SELECT name FROM schema_migrations WHERE batch=$1 ORDER BY name DESC`, last)
    if err != nil {
        return err
    }
    defer rows.Close()
    var batchFiles []migrationFile
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return err
        }
        mf, ok := fileMap[name]
        if !ok {
            return fmt.Errorf("migration file missing for %s", name)
        }
        batchFiles = append(batchFiles, mf)
    }
    if rows.Err() != nil {
        return rows.Err()
    }
    if len(batchFiles) == 0 {
        fmt.Fprintln(cmd.OutOrStdout(), "Nothing to rollback.")
        return nil
    }
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    for _, m := range batchFiles {
        if strings.TrimSpace(m.DownSQL) == "" {
            return fmt.Errorf("migration %s has no down SQL", m.Name)
        }
        if _, err := tx.ExecContext(ctx, m.DownSQL); err != nil {
            return fmt.Errorf("rollback %s: %w", m.Name, err)
        }
        if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE name=$1`, m.Name); err != nil {
            return err
        }
        fmt.Fprintf(cmd.OutOrStdout(), "Rolled back %s\n", m.Name)
    }
    if err := tx.Commit(); err != nil {
        return err
    }
    fmt.Fprintf(cmd.OutOrStdout(), "Rolled back %d migrations from batch %d\n", len(batchFiles), last)
    return nil
}

func migrateStatus(ctx context.Context, cmd *cobra.Command, db *sql.DB, dir string) error {
    files, err := readMigrations(dir)
    if err != nil {
        return err
    }
    applied, err := getApplied(ctx, db)
    if err != nil {
        return err
    }
    if len(files) == 0 {
        fmt.Fprintln(cmd.OutOrStdout(), "No migration files found.")
        return nil
    }
    fmt.Fprintln(cmd.OutOrStdout(), "Name\tStatus")
    for _, f := range files {
        st := "pending"
        if applied[f.Name] {
            st = "applied"
        }
        fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", f.Name, st)
    }
    return nil
}

func getApplied(ctx context.Context, db *sql.DB) (map[string]bool, error) {
    rows, err := db.QueryContext(ctx, `SELECT name FROM schema_migrations`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    out := make(map[string]bool)
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return nil, err
        }
        out[name] = true
    }
    return out, rows.Err()
}

func nextBatch(ctx context.Context, db *sql.DB) (int, error) {
    var max sql.NullInt64
    if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(batch),0) FROM schema_migrations`).Scan(&max); err != nil {
        return 0, err
    }
    return int(max.Int64) + 1, nil
}

func lastBatch(ctx context.Context, db *sql.DB) (int, error) {
    var max sql.NullInt64
    if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(batch),0) FROM schema_migrations`).Scan(&max); err != nil {
        return 0, err
    }
    return int(max.Int64), nil
}

func isPostgres(dsn string) bool {
    l := strings.ToLower(dsn)
    return strings.HasPrefix(l, "postgres://") || strings.HasPrefix(l, "postgresql://")
}

func redactDSN(dsn string) string {
    // very light redaction: strip password if in user:pass@ form
    // postgres://user:pass@host/db -> postgres://user:***@host/db
    if i := strings.Index(dsn, "://"); i != -1 {
        scheme := dsn[:i+3]
        rest := dsn[i+3:]
        at := strings.Index(rest, "@")
        if at != -1 {
            creds := rest[:at]
            if colon := strings.Index(creds, ":"); colon != -1 {
                creds = creds[:colon] + ":***"
                rest = creds + rest[at:]
                return scheme + rest
            }
        }
    }
    return dsn
}


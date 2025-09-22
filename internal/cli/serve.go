package cli

import (
    "context"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "path/filepath"
    "runtime"
    "strings"
    "syscall"
    "time"

    "github.com/spf13/cobra"
)

type serveOptions struct {
    Port int
    Env  string
}

func newServeCmd() *cobra.Command {
    opts := serveOptions{Port: 8080, Env: "dev"}
    cmd := &cobra.Command{
        Use:   "serve [target]",
        Short: "Run the app (dev server)",
        Long:  "Run the generated application's server using 'go run'. Defaults to ./cmd/server. Example: 'largo serve server'",
        Args:  cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            target := "./cmd/server"
            if len(args) == 1 {
                // If user gives a bare name like "server", map to ./cmd/server
                if !strings.Contains(args[0], "/") && !strings.HasPrefix(args[0], ".") {
                    target = filepath.Join("./cmd", args[0])
                } else {
                    target = args[0]
                }
            }
            return runServe(cmd, target, opts)
        },
        Example: "  largo serve\n  largo serve server -p 8080 --env dev\n  largo serve ./cmd/admin",
    }
    cmd.Flags().IntVarP(&opts.Port, "port", "p", opts.Port, "Port to bind (sets PORT env)")
    cmd.Flags().StringVar(&opts.Env, "env", opts.Env, "Environment (sets LARGO_ENV)")
    return cmd
}

func runServe(cmd *cobra.Command, target string, opts serveOptions) error {
    if _, err := exec.LookPath("go"); err != nil {
        return errors.New("'go' tool not found in PATH; install Go or update PATH")
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle Ctrl+C / SIGTERM
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-sigCh
        cancel()
    }()

    args := []string{"run", target}
    c := exec.CommandContext(ctx, "go", args...)
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    c.Stdin = os.Stdin

    // Propagate current env and set PORT/LARGO_ENV if provided
    env := os.Environ()
    if opts.Port > 0 {
        env = upsertEnv(env, "PORT", fmt.Sprintf("%d", opts.Port))
    }
    if opts.Env != "" {
        env = upsertEnv(env, "LARGO_ENV", opts.Env)
    }
    c.Env = env

    start := time.Now()
    fmt.Fprintf(cmd.OutOrStdout(), "Starting dev server: go %s\n", strings.Join(args, " "))
    fmt.Fprintf(cmd.OutOrStdout(), "Env: PORT=%d LARGO_ENV=%s Go=%s %s/%s\n", opts.Port, valueOr(opts.Env, "dev"), runtime.Version(), runtime.GOOS, runtime.GOARCH)

    if err := c.Run(); err != nil {
        // If context was canceled due to signal, treat as clean exit
        if ctx.Err() == context.Canceled {
            fmt.Fprintln(cmd.OutOrStdout(), "Shutting down...")
            return nil
        }
        return err
    }
    fmt.Fprintf(cmd.OutOrStdout(), "Server exited after %s\n", time.Since(start).Round(time.Millisecond))
    return nil
}

func upsertEnv(env []string, key, value string) []string {
    prefix := key + "="
    for i, kv := range env {
        if strings.HasPrefix(kv, prefix) {
            env[i] = prefix + value
            return env
        }
    }
    return append(env, prefix+value)
}

func valueOr(s, def string) string {
    if s == "" {
        return def
    }
    return s
}


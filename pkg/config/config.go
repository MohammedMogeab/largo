package config

import (
    "log/slog"
    "os"
    "strconv"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    App  AppConfig
    DB   DBConfig
    HTTP HTTPConfig
}

type AppConfig struct {
    Env  string
    Port int
}

type DBConfig struct {
    URL string
}

type HTTPConfig struct {
    ReadTimeoutSec  int
    WriteTimeoutSec int
    IdleTimeoutSec  int
    MaxHeaderBytes  int
}

// Load reads .env (if present), then environment variables, then applies defaults.
func Load() *Config {
    _ = godotenv.Load()
    cfg := &Config{}

    // App
    cfg.App.Env = getenvDefault("LARGO_ENV", "dev")
    cfg.App.Port = atoiDefault("PORT", 8080)

    // DB
    cfg.DB.URL = os.Getenv("DATABASE_URL")

    // HTTP
    cfg.HTTP.ReadTimeoutSec = atoiDefault("HTTP_READ_TIMEOUT", 5)
    cfg.HTTP.WriteTimeoutSec = atoiDefault("HTTP_WRITE_TIMEOUT", 10)
    cfg.HTTP.IdleTimeoutSec = atoiDefault("HTTP_IDLE_TIMEOUT", 60)
    cfg.HTTP.MaxHeaderBytes = atoiDefault("HTTP_MAX_HEADER_BYTES", 1<<20) // 1MB

    // Log effective config (concise)
    slog.Info("config.loaded",
        slog.String("env", cfg.App.Env),
        slog.Int("port", cfg.App.Port),
        slog.Int("http_read_timeout_sec", cfg.HTTP.ReadTimeoutSec),
        slog.Int("http_write_timeout_sec", cfg.HTTP.WriteTimeoutSec),
        slog.Int("http_idle_timeout_sec", cfg.HTTP.IdleTimeoutSec),
        slog.Int("http_max_header_bytes", cfg.HTTP.MaxHeaderBytes),
    )
    if cfg.DB.URL == "" {
        slog.Warn("config.database_url_missing", slog.String("hint", "DATABASE_URL required for migrations"))
    }
    return cfg
}

func (h HTTPConfig) ReadTimeout() time.Duration  { return time.Duration(h.ReadTimeoutSec) * time.Second }
func (h HTTPConfig) WriteTimeout() time.Duration { return time.Duration(h.WriteTimeoutSec) * time.Second }
func (h HTTPConfig) IdleTimeout() time.Duration  { return time.Duration(h.IdleTimeoutSec) * time.Second }

func getenvDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func atoiDefault(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    return def
}


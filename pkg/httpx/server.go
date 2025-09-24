package httpx

import (
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/MohammedMogeab/largo/pkg/config"
)

// Serve starts an HTTP server with sane timeouts.
func Serve(addr string, h http.Handler) error {
    srv := &http.Server{
        Addr:              addr,
        Handler:           h,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       5 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
        MaxHeaderBytes:    1 << 20, // 1MB
    }
    return srv.ListenAndServe()
}

// ServeEnv reads PORT from env (default 8080) and serves the handler.
func ServeEnv(h http.Handler) error {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    return Serve(fmt.Sprintf(":%s", port), h)
}

// ServeConfig starts an HTTP server using values from Config.
func ServeConfig(cfg *config.Config, h http.Handler) error {
    addr := fmt.Sprintf(":%d", cfg.App.Port)
    srv := &http.Server{
        Addr:              addr,
        Handler:           h,
        ReadHeaderTimeout: cfg.HTTP.ReadTimeout(),
        ReadTimeout:       cfg.HTTP.ReadTimeout(),
        WriteTimeout:      cfg.HTTP.WriteTimeout(),
        IdleTimeout:       cfg.HTTP.IdleTimeout(),
        MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
    }
    return srv.ListenAndServe()
}

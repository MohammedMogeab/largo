package httpx

import (
    "encoding/json"
    "log/slog"
    "net/http"
    
    "github.com/go-chi/chi/v5"
)

// Context carries request-scoped state and helpers.
type Context struct {
    W         http.ResponseWriter
    R         *http.Request
    Logger    *slog.Logger
    RequestID string
    Values    map[string]any
}

// JSON writes a JSON response with status code.
func (c *Context) JSON(status int, v any) {
    c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.W.WriteHeader(status)
    _ = json.NewEncoder(c.W).Encode(v)
}

// Text writes a text/plain response.
func (c *Context) Text(status int, s string) {
    c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
    c.W.WriteHeader(status)
    _, _ = c.W.Write([]byte(s))
}

// Error writes a JSON error response.
func (c *Context) Error(status int, message string) {
    c.JSON(status, map[string]any{"error": message})
}

// Param returns a URL param captured by the router.
func (c *Context) Param(name string) string {
    return chi.URLParam(c.R, name)
}

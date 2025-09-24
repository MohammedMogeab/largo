package httpx

import (
    "log/slog"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/MohammedMogeab/largo/pkg/httpx/xerr"
)

// HandlerFunc is the httpx handler signature.
type HandlerFunc func(*Context)

// Middleware composes around HandlerFunc.
type Middleware func(HandlerFunc) HandlerFunc

// Router uses chi under the hood and supports param routes.
type Router struct {
    mux         *chi.Mux
    middlewares []Middleware
    logger      *slog.Logger
}

// New creates a Router with sensible defaults and JSON 404/405.
func New() *Router {
    m := chi.NewRouter()
    r := &Router{mux: m, logger: slog.Default()}
    // JSON 404
    m.NotFound(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        rid := req.Header.Get("X-Request-ID")
        xerr.NotFound(w, rid, "route not found")
    }))
    // JSON 405
    m.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        rid := req.Header.Get("X-Request-ID")
        xerr.MethodNotAllowed(w, rid, "")
    }))
    return r
}

// Use appends middleware to the chain.
func (r *Router) Use(mw ...Middleware) { r.middlewares = append(r.middlewares, mw...) }

// Handle registers a route for method and path (supports chi params e.g., /users/{id}).
func (r *Router) Handle(method, path string, h HandlerFunc) {
    r.mux.Method(method, path, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        // Build context
        ctx := &Context{W: w, R: req, Logger: r.logger}
        // Compose chain
        final := h
        for i := len(r.middlewares) - 1; i >= 0; i-- {
            final = r.middlewares[i](final)
        }
        final(ctx)
    }))
}

// GET registers a GET route.
func (r *Router) GET(path string, h HandlerFunc)    { r.Handle(http.MethodGet, path, h) }
// POST registers a POST route.
func (r *Router) POST(path string, h HandlerFunc)   { r.Handle(http.MethodPost, path, h) }
// PUT registers a PUT route.
func (r *Router) PUT(path string, h HandlerFunc)    { r.Handle(http.MethodPut, path, h) }
// DELETE registers a DELETE route.
func (r *Router) DELETE(path string, h HandlerFunc) { r.Handle(http.MethodDelete, path, h) }

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) { r.mux.ServeHTTP(w, req) }

package httpx

import (
    "bytes"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "log/slog"
    "net/http"
    "runtime/debug"
    "time"
)

// RequestID sets a request ID header and stores it in Context.
func RequestID() Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(c *Context) {
            rid := c.R.Header.Get("X-Request-ID")
            if rid == "" {
                rid = newID()
            }
            c.RequestID = rid
            c.W.Header().Set("X-Request-ID", rid)
            next(c)
        }
    }
}

// Logger logs request method, path, status, duration, and request id.
func Logger() Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(c *Context) {
            start := time.Now()
            sr := &statusRecorder{ResponseWriter: c.W}
            c.W = sr
            next(c)
            dur := time.Since(start)
            if sr.status == 0 { // default if handler never wrote
                sr.status = http.StatusOK
            }
            attrs := []slog.Attr{
                slog.String("method", c.R.Method),
                slog.String("path", c.R.URL.Path),
                slog.Int("status", sr.status),
                slog.Int("bytes", sr.n),
                slog.Duration("duration", dur),
            }
            if c.RequestID != "" {
                attrs = append(attrs, slog.String("request_id", c.RequestID))
            }
            c.Logger.Info("http_request", attrs...)
        }
    }
}

// Recover catches panics and returns 500 with a JSON body.
func Recover() Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(c *Context) {
            defer func() {
                if r := recover(); r != nil {
                    // Log stack
                    var b bytes.Buffer
                    fmt.Fprintf(&b, "panic: %v\n", r)
                    b.Write(debug.Stack())
                    c.Logger.Error("panic", slog.String("stack", b.String()))
                    // Respond via unified envelope
                    reqID := c.RequestID
                    // write internal error
                    // local import to avoid circulars
                    httpxInternal(c.W, reqID)
                }
            }()
            next(c)
        }
    }
}

func newID() string {
    var b [12]byte
    if _, err := rand.Read(b[:]); err != nil {
        return fmt.Sprintf("%d", time.Now().UnixNano())
    }
    return hex.EncodeToString(b[:])
}

// indirection to avoid importing xerr at top level and risk cycles
func httpxInternal(w http.ResponseWriter, reqID string) {
    // inline minimal writer to avoid import cycle if packages change
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusInternalServerError)
    _, _ = w.Write([]byte(`{"error":"internal","message":"Internal Server Error","request_id":"` + reqID + `"}`))
}

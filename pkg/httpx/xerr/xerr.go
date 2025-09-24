package xerr

import (
    "encoding/json"
    "net/http"
)

// Envelope is the unified error shape for HTTP errors.
type Envelope struct {
    Error     string      `json:"error"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
    RequestID string      `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

// Helpers for common errors
func BadRequest(w http.ResponseWriter, reqID, msg string, details any) {
    writeJSON(w, http.StatusBadRequest, Envelope{Error: "bad_request", Message: nz(msg, http.StatusText(http.StatusBadRequest)), Details: details, RequestID: reqID})
}

func Unauthorized(w http.ResponseWriter, reqID, msg string) {
    writeJSON(w, http.StatusUnauthorized, Envelope{Error: "unauthorized", Message: nz(msg, http.StatusText(http.StatusUnauthorized)), RequestID: reqID})
}

func Forbidden(w http.ResponseWriter, reqID, msg string) {
    writeJSON(w, http.StatusForbidden, Envelope{Error: "forbidden", Message: nz(msg, http.StatusText(http.StatusForbidden)), RequestID: reqID})
}

func NotFound(w http.ResponseWriter, reqID, msg string) {
    writeJSON(w, http.StatusNotFound, Envelope{Error: "not_found", Message: nz(msg, http.StatusText(http.StatusNotFound)), RequestID: reqID})
}

func MethodNotAllowed(w http.ResponseWriter, reqID, msg string) {
    // 405 mapping (not listed in minimal set) -> use explicit code
    writeJSON(w, http.StatusMethodNotAllowed, Envelope{Error: "method_not_allowed", Message: nz(msg, http.StatusText(http.StatusMethodNotAllowed)), RequestID: reqID})
}

func ValidationFailed(w http.ResponseWriter, reqID string, fields map[string]string) {
    writeJSON(w, http.StatusUnprocessableEntity, Envelope{Error: "validation_failed", Message: "validation failed", Details: fields, RequestID: reqID})
}

func RateLimited(w http.ResponseWriter, reqID, msg string) {
    writeJSON(w, http.StatusTooManyRequests, Envelope{Error: "rate_limited", Message: nz(msg, http.StatusText(http.StatusTooManyRequests)), RequestID: reqID})
}

func Internal(w http.ResponseWriter, reqID, msg string) {
    writeJSON(w, http.StatusInternalServerError, Envelope{Error: "internal", Message: nz(msg, http.StatusText(http.StatusInternalServerError)), RequestID: reqID})
}

func nz(s, def string) string {
    if s == "" { return def }
    return s
}


package httpx

import "net/http"

// statusRecorder wraps ResponseWriter to record status and bytes written.
type statusRecorder struct {
    http.ResponseWriter
    status int
    n      int
}

func (w *statusRecorder) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
    if w.status == 0 {
        w.status = http.StatusOK
    }
    n, err := w.ResponseWriter.Write(b)
    w.n += n
    return n, err
}


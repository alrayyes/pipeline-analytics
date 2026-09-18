package httpserver

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to capture the status code a
// handler wrote, since net/http exposes no way to read it back afterward.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// requestLogger wraps next with a middleware that emits one structured log
// record per request after it completes -- method, path, status, duration,
// remote address -- at a level chosen by the response's status class (5xx
// Error, 4xx Warn, else Info). /healthz is excluded outright: it's polled
// continuously by whatever checks liveness and adds no actionable signal.
//
// Wrapped outside requireSession (not the mux directly), so a request
// requireSession rejects before any handler runs is still logged, with the
// 401/403 status the rejection produced.
//
// Never logs the Cookie or Authorization header, the raw query string, or
// request/response bodies -- only the fields listed above.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)

			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", duration,
			"remoteAddr", r.RemoteAddr,
		}

		switch {
		case rec.status >= http.StatusInternalServerError:
			slog.ErrorContext(r.Context(), "request", attrs...)
		case rec.status >= http.StatusBadRequest:
			slog.WarnContext(r.Context(), "request", attrs...)
		default:
			slog.InfoContext(r.Context(), "request", attrs...)
		}
	})
}

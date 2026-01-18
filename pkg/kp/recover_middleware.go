package kp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/logger"
	"runtime/debug"
	"time"
)

// RecoverMiddleware catches panics during request handling and returns 500.
// It tries to log via logger found in request context; if unavailable, it just responds.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			if rec := recover(); rec != nil {
				var err error
				if e, ok := rec.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", rec)
				}

				// Attempt to log using request-scoped logger
				if lg, ok := r.Context().Value(logger.LoggerKey).(logger.ICustomLogger); ok && lg != nil {
					lg.Error(
						logAction.EXCEPTION("panic recovered"),
						map[string]any{
							"method":   r.Method,
							"path":     r.URL.Path,
							"panic":    err.Error(),
							"duration": time.Since(start).Milliseconds(),
							"stack":    string(debug.Stack()),
						},
					)
					lg.FlushError(http.StatusInternalServerError, "internal_server_error")
				}

				// Respond 500 JSON
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "internal_server_error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

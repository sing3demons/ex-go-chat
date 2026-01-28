package middleware

import (
	"fmt"
	"net/http"
	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/mlog"
	"runtime/debug"
)

// Recovery middleware จัดการ panic ที่เกิดขึ้นใน handler
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				fmt.Printf("Panic recovered: %v\n", err)
				fmt.Printf("Stack trace:\n%s\n", debug.Stack())
				log := mlog.L(r.Context())
				log.Error(logAction.SYSTEM("Panic recovered"), map[string]any{
					"error":      err,
					"stackTrace": string(debug.Stack()),
					"method":     r.Method,
					"url":        r.URL.String(),
					"remoteAddr": r.RemoteAddr,
					"userAgent":  r.UserAgent(),
				})

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal Server Error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

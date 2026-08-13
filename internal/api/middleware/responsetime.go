package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	fmt.Println("ResponseTime")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		duration := time.Since(start)
		wrappedWriter.Header().Set("X-Response-Time", duration.String())

		next.ServeHTTP(wrappedWriter, r)
		duration = time.Since(start)

	})

}

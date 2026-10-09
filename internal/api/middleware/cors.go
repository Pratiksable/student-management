package middleware

import (
	"net/http"
)

var allowedOrigins = []string{
	"http://localhost:3000",
	"https://localhost:3000",
	"https://my-origin-website.com",
}

func CORS(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			return
		}
		origin := r.Header.Get("Origin")

		// Non-browser API clients (for example Insomnia and curl) do not send
		// Origin. CORS does not apply to those requests, so let them through.
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if isOriginALlowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

		} else {
			http.Error(w, "Not Allowed by the cors", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isOriginALlowed(origin string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

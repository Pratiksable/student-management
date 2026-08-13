package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

type ZgripResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (g *ZgripResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

func CompressionMiddleware(next http.Handler) http.Handler {
	fmt.Println("Compression")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w = &ZgripResponseWriter{ResponseWriter: w, Writer: gz}

		next.ServeHTTP(w, r)

	})

}

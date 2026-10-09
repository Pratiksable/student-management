package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Pratiksable/student-management/internal/api/middleware"
	"github.com/Pratiksable/student-management/internal/api/router"
	sqlconnect "github.com/Pratiksable/student-management/internal/repositories/sql-connect"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}

	var PORT = os.Getenv("SERVER_PORT")

	_, error := sqlconnect.ConnectDB()
	if error != nil {
		fmt.Println("Error: ", error)
		return
	}

	// TLS certificate files
	cert := "cert.pem"
	key := "key.pem"

	// -----------------------------
	// Rate limiter
	// -----------------------------

	rateLimiter := middleware.NewRateLimiter(
		6,
		time.Minute,
	)

	// -----------------------------
	// HPP configuration
	// -----------------------------

	hppOptions := middleware.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",

		Whitelist: []string{
			"sortby",
			"sortOrder",
			"first_name",
			"last_name",
			"age",
			"class",
			"subject",
			"email",
		},
	}

	// -----------------------------
	// Middleware chain
	// -----------------------------

	var handler http.Handler = router.MainRouter()

	// Closest to route handler
	handler = middleware.CompressionMiddleware(handler)

	// Clean query/body parameters
	handler = middleware.Hpp(hppOptions)(handler)

	// Reject users exceeding request limit
	handler = rateLimiter.RateLimitingMiddleware(handler)

	// Measure request processing time
	handler = middleware.ResponseTimeMiddleware(handler)

	// Handle CORS
	handler = middleware.CORS(handler)

	// Add security headers
	handler = middleware.SecurityHeaders(handler)

	// -----------------------------
	// TLS configuration
	// -----------------------------

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// -----------------------------
	// HTTP server
	// -----------------------------

	server := &http.Server{
		Addr:      ":" + PORT,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port", PORT)

	err = server.ListenAndServeTLS(cert, key)

	if err != nil {
		log.Fatal("Error starting the server: ", err)
	}
}

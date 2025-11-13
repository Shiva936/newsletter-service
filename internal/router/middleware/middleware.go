package middleware

import (
	"log"
	"net/http"
	"time"
)

// BasicAuth middleware protects routes with given username and password
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized\n"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs incoming requests and their responses including duration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Use a ResponseWriter wrapper to capture status code
		lrw := &loggingResponseWriter{w, http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %s\n", r.Method, r.RequestURI, lrw.statusCode, duration)
		
        //TODO: Implemnet api metrices push to a bucket for monitoring
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code for logging
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

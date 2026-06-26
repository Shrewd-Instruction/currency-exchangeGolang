package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"currency-exchange/logger"
	"currency-exchange/models"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		logger.Log.Info().Msgf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type visitor struct {
	count    int
	lastSeen time.Time
	mu       sync.Mutex
}

func rateLimiter(maxReqs int, window time.Duration) func(http.Handler) http.Handler {
	var visitors sync.Map

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			val, _ := visitors.LoadOrStore(ip, &visitor{})
			v := val.(*visitor)

			v.mu.Lock()
			defer v.mu.Unlock()
			if time.Since(v.lastSeen) > window {
				v.count = 0
				v.lastSeen = time.Now()
			}
			v.count++
			if v.count > maxReqs {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(models.APIError{
					Code:    429,
					Message: fmt.Sprintf("rate limit exceeded, max %d requests per %v", maxReqs, window),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

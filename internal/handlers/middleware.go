package handlers

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		defer cfg.mu.Unlock()
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) MiddlewareMetricsResult(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type", "text/html")
	writer.Write([]byte(body))
}

func (cfg *ApiConfig) MiddlewareMetricsReset(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.fileserverHits = 0
	writer.WriteHeader(http.StatusOK)
}

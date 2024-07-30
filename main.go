package main

import (
	"fmt"
	"net/http"
	"sync"
)

type apiConfig struct {
	mu             sync.Mutex
	fileserverHits int
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}
	mux.Handle("/app/*", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.middlewareMetricsResult)
	mux.HandleFunc("/api/reset", apiCfg.middlewareMetricsReset)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: mux,
	}
	fmt.Println(server.ListenAndServe())
}

func readinessHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.mu.Lock()
		defer cfg.mu.Unlock()
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsResult(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)
	writer.WriteHeader(200)
	writer.Header().Add("Content-Type", "text/html")
	writer.Write([]byte(body))
}

func (cfg *apiConfig) middlewareMetricsReset(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.fileserverHits = 0
	writer.WriteHeader(200)
}

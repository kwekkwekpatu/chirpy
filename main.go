package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kwekkwekpatu/chirpy/internal/handlers"
	_ "github.com/lib/pq"
)

func main() {
	mux := http.NewServeMux()
	apiCfg := handlers.APIConfig

	port := os.Getenv("PORT")
	port = ":" + port

	mux.Handle("GET /app/*", http.StripPrefix("/app", apiCfg.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlers.ReadinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.MiddlewareMetricsResult)
	mux.HandleFunc("GET /api/chirps", apiCfg.ChirpReadHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.ChirpSpecificReadHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.ChirpDeleteSpecificHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.MiddlewareMetricsReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.ChirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.UserHandler)
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.AdminReset)
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUserPasswordHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.UpgradeUser)

	mux.HandleFunc("GET /", handlers.DockerHandler)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	fmt.Println(server.ListenAndServe())
}

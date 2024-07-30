package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/*", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", readinessHandler)
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

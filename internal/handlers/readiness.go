package handlers

import (
	"net/http"

	_ "github.com/lib/pq"
)

func ReadinessHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

package util

import (
	"encoding/json"
	"net/http"

	_ "github.com/lib/pq"
)

var InternalServerError = "Something went wrong"

func RespondWithError(writer http.ResponseWriter, request *http.Request, code int, message string, err error) {
	if err != nil {
		ErrorLogger.Println(err)
	}
	if code > 499 {
		ErrorLogger.Printf("Responding with 5XX error: %s", message)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	RespondWithJson(writer, request, code, errorResponse{
		Error: message,
	})
}

func RespondWithJson(writer http.ResponseWriter, request *http.Request, code int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		ErrorLogger.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(`{"error":"Failed to Marshal response."}`))
		return
	}

	writer.WriteHeader(code)
	writer.Write(dat)
}

package handlers

import (
	"net/http"
	"os"

	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

func DockerHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}
	filePath := "/public/index.html"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		util.ErrorLogger.Printf("File not found: %s", filePath)
		http.Error(writer, "File not found", http.StatusNotFound)
		return
	}
	http.ServeFile(writer, request, filePath)
}

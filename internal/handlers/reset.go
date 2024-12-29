package handlers

import (
	"net/http"

	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

func (cfg *ApiConfig) AdminReset(writer http.ResponseWriter, request *http.Request) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.platform != "dev" {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	err := cfg.db.DeleteUsers(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to delete users.", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	return
}

package handlers

import (
	"net/http"

	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

func (cfg *ApiConfig) AdminReset(writer http.ResponseWriter, request *http.Request) {
	util.InfoLogger.Println("Handling admin reset")
	if cfg.platform != "dev" {
		util.ErrorLogger.Println("The current environment is not a dev environment. Admin reset not allowed")
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	util.InfoLogger.Println("The current environment is a dev environment. Admin reset is allowed")
	err := cfg.db.DeleteUsers(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to delete users.", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	util.InfoLogger.Println("Succesfully performed admin reset.")
	return
}

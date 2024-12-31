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
	util.InfoLogger.Println("Deleting users")
	err := cfg.db.DeleteUsers(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to delete users.", err)
		return
	}
	util.InfoLogger.Println("Users have been deleted")

	util.InfoLogger.Println("Deleting chirps")
	err = cfg.db.DeleteChirp(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to delete chirps.", err)
		return
	}
	util.InfoLogger.Println("Chirps have been deleted")

	util.InfoLogger.Println("Deleting refresh tokens")
	err = cfg.db.DeleteRefreshTokens(request.Context())
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to delete refresh tokens.", err)
		return
	}
	util.InfoLogger.Println("Refresh tokens have been deleted")

	writer.WriteHeader(http.StatusOK)
	util.InfoLogger.Println("Succesfully performed admin reset.")
	return
}

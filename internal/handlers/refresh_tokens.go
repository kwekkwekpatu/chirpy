package handlers

import (
	"net/http"
	"time"

	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/util"
)

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *ApiConfig) RefreshHandler(writer http.ResponseWriter, request *http.Request) {
	util.InfoLogger.Printf("Handling refresh.")
	util.InfoLogger.Printf("Extracting and validating Refresh token")

	tokenString, err := auth.GetBearerToken(request.Header)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Missing authorization header.", err)
		return
	}

	dbToken, err := cfg.db.ReadRefreshToken(request.Context(), tokenString)
	if err != nil || (dbToken.RevokedAt.Valid && time.Now().After(dbToken.RevokedAt.Time)) {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "No valid token in database.", err)
		return
	}

	dbUser, err := cfg.db.GetUserFromRefreshToken(request.Context(), tokenString)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Refresh token does not belong to any known users.", err)
		return
	}

	util.InfoLogger.Printf("Generating new token.")
	newToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to generate new token.", err)
		return
	}

	responseBody := RefreshResponse{Token: newToken}
	util.RespondWithJson(writer, request, http.StatusOK, responseBody)
	util.InfoLogger.Printf("Successfully refreshed token.")

}

func (cfg *ApiConfig) RevokeHandler(writer http.ResponseWriter, request *http.Request) {
	util.InfoLogger.Printf("Handling refresh.")
	util.InfoLogger.Printf("Extracting and validating Refresh token")

	tokenString, err := auth.GetBearerToken(request.Header)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Missing authorization header.", err)
		return
	}

	util.InfoLogger.Printf("Revoking old refresh token.")
	err = cfg.db.RevokeRefreshToken(request.Context(), tokenString)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to revoke old refresh token.", err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
	util.InfoLogger.Printf("Revoked old refresh token.")

}

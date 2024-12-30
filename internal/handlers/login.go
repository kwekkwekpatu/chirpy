package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/util"
)

func (cfg *ApiConfig) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	util.InfoLogger.Printf("Handling user login.")

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, util.InternalServerError, err)
		return
	}

	dbUser, err := cfg.db.GetUserByEmail(request.Context(), params.Email)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	util.InfoLogger.Printf("Generating response body from login.")
	responseBody := User{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email}
	util.RespondWithJson(writer, request, http.StatusOK, responseBody)

	util.InfoLogger.Printf("Successfully logged in.")
}

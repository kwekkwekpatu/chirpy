package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/util"
)

const (
	MaxExpirationSeconds = 3600 // 1 hour in seconds
)

type LoginResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *ApiConfig) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
	}

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

	expiresIn := MaxExpirationSeconds

	if params.ExpiresInSeconds != nil {
		// We are capping the expiry timer to the default value.
		// So we only need to do something if the expiry is smaller than default.
		if *params.ExpiresInSeconds <= 0 {
			util.RespondWithError(writer, request, http.StatusBadRequest, "Expiration time must be positive", nil)
			return
		}
		if *params.ExpiresInSeconds < MaxExpirationSeconds {
			expiresIn = *params.ExpiresInSeconds
		}
	}
	duration := time.Duration(expiresIn) * time.Second

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, duration)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to create token.", err)
		return
	}

	util.InfoLogger.Printf("Generating response body from login.")
	responseBody := LoginResponse{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt, UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email, Token: token}

	util.RespondWithJson(writer, request, http.StatusOK, responseBody)

	util.InfoLogger.Printf("Successfully logged in.")
}

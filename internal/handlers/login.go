package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/database"
	"github.com/kwekkwekpatu/chirpy/internal/util"
)

const (
	MaxExpirationSeconds      = 3600           // 1 hour in seconds
	RefreshExpirationDuration = 60 * 24 * 3600 // 60 days in seconds
)

type LoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ChirpyIsRed  bool      `json:"is_chirpy_red"`
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

	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to create token.", err)
		return
	}

	refreshDuration := time.Duration(RefreshExpirationDuration) * time.Second
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to create refresh_token.", err)
		return
	}

	refreshParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(refreshDuration),
		UserID:    dbUser.ID,
	}

	_, err = cfg.db.CreateRefreshToken(request.Context(), refreshParams)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Failed to store refresh_token.", err)
		return
	}

	util.InfoLogger.Printf("Generating response body from login.")
	responseBody := LoginResponse{ID: dbUser.ID, CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt, Email: dbUser.Email, Token: token,
		RefreshToken: refreshToken, ChirpyIsRed: dbUser.IsChirpyRed}

	util.RespondWithJson(writer, request, http.StatusOK, responseBody)

	util.InfoLogger.Printf("Successfully logged in.")
}

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
	"github.com/kwekkwekpatu/chirpy/internal/database"
	"github.com/kwekkwekpatu/chirpy/internal/util"
	_ "github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *ApiConfig) UserHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	util.InfoLogger.Printf("Handling user creation.")

	util.InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, util.InternalServerError, err)
		return
	}

	util.InfoLogger.Printf("Successfully loaded email: %s", params.Email)
	if params.Email == "" {
		util.WarnLogger.Printf("The email paramater is empty. Cannot create a new user without an email.")
		util.RespondWithError(writer, request, http.StatusBadRequest, "Email is required", nil)
		return
	}

	util.InfoLogger.Printf("Processing password")
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to hash password", err)
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	util.InfoLogger.Printf("Attempting to create user with email: %s", params.Email)
	user, err := cfg.db.CreateUser(request.Context(), userParams)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to create user.", err)
		return
	}
	util.InfoLogger.Printf("Successfully created a user for email: %s", params.Email)

	util.InfoLogger.Printf("Generating response body from user.")
	responseBody := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	util.RespondWithJson(writer, request, http.StatusCreated, responseBody)

	util.InfoLogger.Printf("Successfully created user.")
	return
}

func (cfg *ApiConfig) UpdateUserPasswordHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	util.InfoLogger.Printf("Entering password handler.")

	util.InfoLogger.Printf("Extracting and validating JWT token")
	tokenString, err := auth.GetBearerToken(request.Header)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "Missing authorization header.", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusUnauthorized, "JWT is invalid", err)
		return
	}

	util.InfoLogger.Printf("Handling password update.")

	util.InfoLogger.Printf("Loading request parameter.")
	writer.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, util.InternalServerError, err)
		return
	}

	util.InfoLogger.Printf("Successfully loaded email: %s", params.Email)
	if params.Email == "" {
		util.WarnLogger.Printf("The email parameter is empty. Cannot update a user without an email.")
		util.RespondWithError(writer, request, http.StatusBadRequest, "Email is required", nil)
		return
	}

	util.InfoLogger.Printf("Processing new password")
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to hash password", err)
	}

	userParams := database.PutPasswordByUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	util.InfoLogger.Printf("Attempting to update password with email: %s", params.Email)
	user, err := cfg.db.PutPasswordByUser(request.Context(), userParams)
	if err != nil {
		util.RespondWithError(writer, request, http.StatusInternalServerError, "Failed to update password.", err)
		return
	}
	util.InfoLogger.Printf("Successfully updated password for email: %s", params.Email)

	util.InfoLogger.Printf("Generating response body from user.")
	responseBody := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email}
	util.RespondWithJson(writer, request, http.StatusOK, responseBody)

	util.InfoLogger.Printf("Successfully updated password.")
	return
}

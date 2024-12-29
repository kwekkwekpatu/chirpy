package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
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
		Email string `json:"email"`
	}

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
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

	util.InfoLogger.Printf("Attempting to create user with email: %s", params.Email)
	user, err := cfg.db.CreateUser(request.Context(), params.Email)
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
